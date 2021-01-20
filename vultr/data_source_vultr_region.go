package vultr

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vultr/govultr/v2"
)

func dataSourceVultrRegion() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVultrRegionRead,
		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"country": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"continent": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"city": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVultrRegionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).govultrClient()

	filters, filtersOk := d.GetOk("filter")

	if !filtersOk {
		return fmt.Errorf("issue with filter: %v", filtersOk)
	}

	regionList := []govultr.Region{}
	f := buildVultrDataSourceFilter(filters.(*schema.Set))
	options := &govultr.ListOptions{}
	for {
		regions, meta, err := client.Region.List(context.Background(), options)
		if err != nil {
			return fmt.Errorf("Error getting regions: %v", err)
		}

		for _, a := range regions {
			// we need convert the a struct INTO a map so we can easily manipulate the data here
			sm, err := structToMap(a)

			if err != nil {
				return err
			}

			if filterLoop(f, sm) {
				regionList = append(regionList, a)
			}
		}

		if meta.Links.Next == "" {
			break
		} else {
			options.Cursor = meta.Links.Next
			continue
		}
	}

	if len(regionList) > 1 {
		return errors.New("your search returned too many results. Please refine your search to be more specific")
	}

	if len(regionList) < 1 {
		return errors.New("no results were found")
	}

	d.SetId(regionList[0].ID)
	d.Set("country", regionList[0].Country)
	d.Set("continent", regionList[0].Continent)
	d.Set("city", regionList[0].City)
	d.Set("options", regionList[0].Options)
	return nil
}
