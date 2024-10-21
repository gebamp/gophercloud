//go:build acceptance || compute || secgroups

package v2

import (
	"context"
	"testing"

	"github.com/gophercloud/gophercloud/v2/internal/acceptance/clients"
	"github.com/gophercloud/gophercloud/v2/internal/acceptance/tools"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/secgroups"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	th "github.com/gophercloud/gophercloud/v2/testhelper"
)

func TestSecGroupsList(t *testing.T) {
	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	allPages, err := secgroups.List(client).AllPages(context.TODO())
	th.AssertNoErr(t, err)

	allSecGroups, err := secgroups.ExtractSecurityGroups(allPages)
	th.AssertNoErr(t, err)

	var found bool
	for _, secgroup := range allSecGroups {
		tools.PrintResource(t, secgroup)

		if secgroup.Name == "default" {
			found = true
		}
	}

	th.AssertEquals(t, true, found)
}

func TestSecGroupsCRUD(t *testing.T) {
	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	securityGroup, err := CreateSecurityGroup(t, client)
	th.AssertNoErr(t, err)
	defer DeleteSecurityGroup(t, client, securityGroup.ID)

	tools.PrintResource(t, securityGroup)

	newName := tools.RandomString("secgroup_", 4)
	description := ""
	updateOpts := secgroups.UpdateOpts{
		Name:        newName,
		Description: &description,
	}
	updatedSecurityGroup, err := secgroups.Update(context.TODO(), client, securityGroup.ID, updateOpts).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, updatedSecurityGroup)

	t.Logf("Updated %s's name to %s", updatedSecurityGroup.ID, updatedSecurityGroup.Name)

	th.AssertEquals(t, newName, updatedSecurityGroup.Name)
	th.AssertEquals(t, description, updatedSecurityGroup.Description)
}

func TestSecGroupsRuleCreate(t *testing.T) {
	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	securityGroup, err := CreateSecurityGroup(t, client)
	th.AssertNoErr(t, err)
	defer DeleteSecurityGroup(t, client, securityGroup.ID)

	tools.PrintResource(t, securityGroup)

	rule, err := CreateSecurityGroupRule(t, client, securityGroup.ID)
	th.AssertNoErr(t, err)
	defer DeleteSecurityGroupRule(t, client, rule.ID)

	tools.PrintResource(t, rule)

	newSecurityGroup, err := secgroups.Get(context.TODO(), client, securityGroup.ID).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, newSecurityGroup)

	th.AssertEquals(t, 1, len(newSecurityGroup.Rules))
}

func TestSecGroupsAddGroupToServer(t *testing.T) {
	clients.RequireLong(t)

	client, err := clients.NewComputeV2Client()
	th.AssertNoErr(t, err)

	server, err := CreateServer(t, client)
	th.AssertNoErr(t, err)
	defer DeleteServer(t, client, server)

	securityGroup, err := CreateSecurityGroup(t, client)
	th.AssertNoErr(t, err)
	defer DeleteSecurityGroup(t, client, securityGroup.ID)

	rule, err := CreateSecurityGroupRule(t, client, securityGroup.ID)
	th.AssertNoErr(t, err)
	defer DeleteSecurityGroupRule(t, client, rule.ID)

	t.Logf("Adding group %s to server %s", securityGroup.ID, server.ID)
	err = secgroups.AddServer(context.TODO(), client, server.ID, securityGroup.Name).ExtractErr()
	th.AssertNoErr(t, err)

	server, err = servers.Get(context.TODO(), client, server.ID).Extract()
	th.AssertNoErr(t, err)

	tools.PrintResource(t, server)

	var found bool
	for _, sg := range server.SecurityGroups {
		if sg["name"] == securityGroup.Name {
			found = true
		}
	}

	th.AssertEquals(t, true, found)

	t.Logf("Removing group %s from server %s", securityGroup.ID, server.ID)
	err = secgroups.RemoveServer(context.TODO(), client, server.ID, securityGroup.Name).ExtractErr()
	th.AssertNoErr(t, err)

	server, err = servers.Get(context.TODO(), client, server.ID).Extract()
	th.AssertNoErr(t, err)

	found = false

	tools.PrintResource(t, server)

	for _, sg := range server.SecurityGroups {
		if sg["name"] == securityGroup.Name {
			found = true
		}
	}

	th.AssertEquals(t, false, found)
}
