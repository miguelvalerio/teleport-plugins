/*
Copyright 2015-2021 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"github.com/gravitational/trace"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func (s *TerraformSuite) TestOIDCConnector() {
	if !s.teleportFeatures.AdvancedAccessWorkflows {
		s.T().Skip("AdvancedAccessWorkflows are disabled")
	}

	res := "teleport_oidc_connector"

	create := s.terraformConfig + `
		resource "` + res + `" "test" {
			metadata {
				name    = "test"
				expires = "2022-10-12T07:20:50.3Z"
				labels  = {
				  	example = "yes"
				}
			}

			spec {
				client_id = "client"
				client_secret = "value"
			
				claims_to_roles {
					claim = "test"
					roles = ["terraform"]
				}
			}
		}
	`

	update := s.terraformConfig + `
		resource "` + res + `" "test" {
			metadata {
				name    = "test"
				expires = "2022-10-12T07:20:50.3Z"
				labels  = {
					example = "yes"
				}
			}

			spec {
				client_id = "client"
				client_secret = "value"
			
				claims_to_roles {
					claim = "test"
					roles = ["teleport"]
				}
			}
		}
	`

	checkOIDCConnectorDestroyed := func(state *terraform.State) error {
		_, err := s.client.GetOIDCConnector(s.Context(), "test", true)
		if trace.IsNotFound(err) {
			return nil
		}

		return err
	}

	name := res + ".test"

	resource.Test(s.T(), resource.TestCase{
		ProviderFactories: s.terraformProviders,
		CheckDestroy:      checkOIDCConnectorDestroyed,
		Steps: []resource.TestStep{
			{
				Config: create,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "kind", "oidc"),
					resource.TestCheckResourceAttr(name, "metadata.0.expires", "2022-10-12T07:20:50.3Z"),
					resource.TestCheckResourceAttr(name, "spec.0.client_id", "client"),
					resource.TestCheckResourceAttr(name, "spec.0.claims_to_roles.0.claim", "test"),
					resource.TestCheckResourceAttr(name, "spec.0.claims_to_roles.0.roles.0", "terraform"),
				),
			},
			{
				Config:   create, // Check that there is no state drift
				PlanOnly: true,
			},
			{
				Config: update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "kind", "oidc"),
					resource.TestCheckResourceAttr(name, "metadata.0.expires", "2022-10-12T07:20:50.3Z"),
					resource.TestCheckResourceAttr(name, "spec.0.client_id", "client"),
					resource.TestCheckResourceAttr(name, "spec.0.claims_to_roles.0.claim", "test"),
					resource.TestCheckResourceAttr(name, "spec.0.claims_to_roles.0.roles.0", "teleport"),
				),
			},
			{
				Config:   update, // Check that there is no state drift
				PlanOnly: true,
			},
		},
	})
}
