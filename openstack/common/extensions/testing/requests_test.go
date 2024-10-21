package testing

import (
	"context"
	"testing"

	"github.com/gophercloud/gophercloud/v2/openstack/common/extensions"
	"github.com/gophercloud/gophercloud/v2/pagination"
	th "github.com/gophercloud/gophercloud/v2/testhelper"
	"github.com/gophercloud/gophercloud/v2/testhelper/client"
)

func TestList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleListExtensionsSuccessfully(t)

	count := 0

	err := extensions.List(client.ServiceClient()).EachPage(context.TODO(), func(_ context.Context, page pagination.Page) (bool, error) {
		count++
		actual, err := extensions.ExtractExtensions(page)
		th.AssertNoErr(t, err)
		th.AssertDeepEquals(t, ExpectedExtensions, actual)

		return true, nil
	})
	th.AssertNoErr(t, err)

	th.CheckEquals(t, 1, count)
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleGetExtensionSuccessfully(t)

	actual, err := extensions.Get(context.TODO(), client.ServiceClient(), "agent").Extract()
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, SingleExtension, actual)
}
