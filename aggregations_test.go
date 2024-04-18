package prago

import (
	"fmt"
	"testing"
)

func TestAggregations(t *testing.T) {
	resource := prepareResource()
	resA := ResourceStruct{Count: 1}
	resB := ResourceStruct{Count: 2, IsSomething: true}
	resC := ResourceStruct{Count: 3, IsSomething: true}

	CreateItem(resource.app, &resA)
	CreateItem(resource.app, &resB)
	CreateItem(resource.app, &resC)

	res, err := Query[ResourceStruct](resource.app).Is("IsSomething", true).Aggregation().Count().Sum("Count").Min("Count").Max("Count").Get()
	if err != nil {
		t.Fatal(err)
	}

	if fmt.Sprintf("%v", res) != fmt.Sprintf("%v", []int64{2, 5, 2, 3}) {
		t.Fatal(res)
	}
}
