//go:build acceptance || compute || flavors

package v2

import (
	"context"
	"testing"

	"github.com/gophercloud/gophercloud/v2/internal/acceptance/clients"
	identity "github.com/gophercloud/gophercloud/v2/internal/acceptance/openstack/identity/v3"
	"github.com/gophercloud/gophercloud/v2/internal/acceptance/tools"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/flavors"
	th "github.com/gophercloud/gophercloud/v2/testhelper"
)

func TestFlavorsList(t *testing.T) {
	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	choices, err := clients.AcceptanceTestChoicesFromEnv()
	th.AssertNoErr(t, err)

	allPages, err := flavors.ListDetail(client, nil).AllPages(context.TODO())
	th.AssertNoErr(t, err)

	allFlavors, err := flavors.ExtractFlavors(allPages)
	th.AssertNoErr(t, err)

	var found bool
	for _, flavor := range allFlavors {
		tools.PrintResource(t, flavor)

		if flavor.ID == choices.FlavorID {
			found = true
		}
	}

	th.AssertEquals(t, true, found)
}

func TestFlavorsAccessTypeList(t *testing.T) {
	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	flavorAccessTypes := map[string]flavors.AccessType{
		"public":  flavors.PublicAccess,
		"private": flavors.PrivateAccess,
		"all":     flavors.AllAccess,
	}

	for flavorTypeName, flavorAccessType := range flavorAccessTypes {
		t.Logf("** %s flavors: **", flavorTypeName)
		allPages, err := flavors.ListDetail(client, flavors.ListOpts{AccessType: flavorAccessType}).AllPages(context.TODO())
		th.AssertNoErr(t, err)

		allFlavors, err := flavors.ExtractFlavors(allPages)
		th.AssertNoErr(t, err)

		for _, flavor := range allFlavors {
			tools.PrintResource(t, flavor)
		}
	}
}

func TestFlavorsGet(t *testing.T) {
	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	choices, err := clients.AcceptanceTestChoicesFromEnv()
	th.AssertNoErr(t, err)

	flavor, err := flavors.Get(context.TODO(), client, choices.FlavorID).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, flavor)

	th.AssertEquals(t, choices.FlavorID, flavor.ID)
}

func TestFlavorExtraSpecsGet(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	// Microversion 2.61 is required to add extra_specs to flavor
	client.Microversion = "2.61"

	flavor, err := CreatePrivateFlavor(t, client)
	th.AssertNoErr(t, err)
	defer DeleteFlavor(t, client, flavor)

	createOpts := flavors.ExtraSpecsOpts{
		"hw:cpu_policy":        "CPU-POLICY",
		"hw:cpu_thread_policy": "CPU-THREAD-POLICY",
	}
	createdExtraSpecs, err := flavors.CreateExtraSpecs(context.TODO(), client, flavor.ID, createOpts).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, createdExtraSpecs)

	flavor, err = flavors.Get(context.TODO(), client, flavor.ID).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, flavor)
	th.AssertEquals(t, 2, len(flavor.ExtraSpecs))
	th.AssertEquals(t, "CPU-POLICY", flavor.ExtraSpecs["hw:cpu_policy"])
	th.AssertEquals(t, "CPU-THREAD-POLICY", flavor.ExtraSpecs["hw:cpu_thread_policy"])
}

func TestFlavorsCreateDelete(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	flavor, err := CreateFlavor(t, client)
	th.AssertNoErr(t, err)
	defer DeleteFlavor(t, client, flavor)

	tools.PrintResource(t, flavor)
}

func TestFlavorsCreateUpdateDelete(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	flavor, err := CreateFlavor(t, client)
	th.AssertNoErr(t, err)
	defer DeleteFlavor(t, client, flavor)

	tools.PrintResource(t, flavor)

	newFlavorDescription := "This is the new description"
	updateOpts := flavors.UpdateOpts{
		Description: newFlavorDescription,
	}

	flavor, err = flavors.Update(context.TODO(), client, flavor.ID, updateOpts).Extract()
	th.AssertNoErr(t, err)
	th.AssertEquals(t, newFlavorDescription, flavor.Description)

	tools.PrintResource(t, flavor)
}

func TestFlavorsAccessesList(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	flavor, err := CreatePrivateFlavor(t, client)
	th.AssertNoErr(t, err)
	defer DeleteFlavor(t, client, flavor)

	allPages, err := flavors.ListAccesses(client, flavor.ID).AllPages(context.TODO())
	th.AssertNoErr(t, err)

	allAccesses, err := flavors.ExtractAccesses(allPages)
	th.AssertNoErr(t, err)

	th.AssertEquals(t, 0, len(allAccesses))
}

func TestFlavorsAccessCRUD(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	identityClient, err := clients.NewIdentityV3Client()
	th.AssertNoErr(t, err)

	project, err := identity.CreateProject(t, identityClient, nil)
	th.AssertNoErr(t, err)
	defer identity.DeleteProject(t, identityClient, project.ID)

	flavor, err := CreatePrivateFlavor(t, client)
	th.AssertNoErr(t, err)
	defer DeleteFlavor(t, client, flavor)

	addAccessOpts := flavors.AddAccessOpts{
		Tenant: project.ID,
	}

	accessList, err := flavors.AddAccess(context.TODO(), client, flavor.ID, addAccessOpts).Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, 1, len(accessList))
	th.AssertEquals(t, project.ID, accessList[0].TenantID)
	th.AssertEquals(t, flavor.ID, accessList[0].FlavorID)

	for _, access := range accessList {
		tools.PrintResource(t, access)
	}

	removeAccessOpts := flavors.RemoveAccessOpts{
		Tenant: project.ID,
	}

	accessList, err = flavors.RemoveAccess(context.TODO(), client, flavor.ID, removeAccessOpts).Extract()
	th.AssertNoErr(t, err)

	th.AssertEquals(t, 0, len(accessList))
}

func TestFlavorsExtraSpecsCRUD(t *testing.T) {
	clients.RequireAdmin(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	flavor, err := CreatePrivateFlavor(t, client)
	th.AssertNoErr(t, err)
	defer DeleteFlavor(t, client, flavor)

	createOpts := flavors.ExtraSpecsOpts{
		"hw:cpu_policy":        "CPU-POLICY",
		"hw:cpu_thread_policy": "CPU-THREAD-POLICY",
	}
	createdExtraSpecs, err := flavors.CreateExtraSpecs(context.TODO(), client, flavor.ID, createOpts).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, createdExtraSpecs)

	th.AssertEquals(t, 2, len(createdExtraSpecs))
	th.AssertEquals(t, "CPU-POLICY", createdExtraSpecs["hw:cpu_policy"])
	th.AssertEquals(t, "CPU-THREAD-POLICY", createdExtraSpecs["hw:cpu_thread_policy"])

	err = flavors.DeleteExtraSpec(context.TODO(), client, flavor.ID, "hw:cpu_policy").ExtractErr()
	th.AssertNoErr(t, err)

	updateOpts := flavors.ExtraSpecsOpts{
		"hw:cpu_thread_policy": "CPU-THREAD-POLICY-BETTER",
	}
	updatedExtraSpec, err := flavors.UpdateExtraSpec(context.TODO(), client, flavor.ID, updateOpts).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, updatedExtraSpec)

	allExtraSpecs, err := flavors.ListExtraSpecs(context.TODO(), client, flavor.ID).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, allExtraSpecs)

	th.AssertEquals(t, 1, len(allExtraSpecs))
	th.AssertEquals(t, "CPU-THREAD-POLICY-BETTER", allExtraSpecs["hw:cpu_thread_policy"])

	spec, err := flavors.GetExtraSpec(context.TODO(), client, flavor.ID, "hw:cpu_thread_policy").Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, spec)

	th.AssertEquals(t, "CPU-THREAD-POLICY-BETTER", spec["hw:cpu_thread_policy"])
}
