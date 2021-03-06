package digitalocean

import (
	"context"
	"fmt"
	"testing"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceDigitalOceanLoadBalancer_Basic(t *testing.T) {
	var loadbalancer godo.LoadBalancer
	rInt := acctest.RandInt()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDataSourceDigitalOceanLoadBalancerConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataSourceDigitalOceanLoadBalancerExists("data.digitalocean_loadbalancer.foobar", &loadbalancer),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "name", fmt.Sprintf("loadbalancer-%d", rInt)),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "region", "nyc3"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "forwarding_rule.#", "1"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "forwarding_rule.0.entry_port", "80"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "forwarding_rule.0.entry_protocol", "http"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "forwarding_rule.0.target_port", "80"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "forwarding_rule.0.target_protocol", "http"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "healthcheck.#", "1"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "healthcheck.0.port", "22"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "healthcheck.0.protocol", "tcp"),
					resource.TestCheckResourceAttr(
						"data.digitalocean_loadbalancer.foobar", "droplet_ids.#", "2"),
				),
			},
		},
	})
}

func testAccCheckDataSourceDigitalOceanLoadBalancerExists(n string, loadbalancer *godo.LoadBalancer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Load Balancer ID is set")
		}

		client := testAccProvider.Meta().(*godo.Client)

		foundLoadbalancer, _, err := client.LoadBalancers.Get(context.Background(), rs.Primary.ID)

		if err != nil {
			return err
		}

		if foundLoadbalancer.ID != rs.Primary.ID {
			return fmt.Errorf("Load Balancer not found")
		}

		*loadbalancer = *foundLoadbalancer

		return nil
	}
}

func testAccCheckDataSourceDigitalOceanLoadBalancerConfig_basic(rInt int) string {
	return fmt.Sprintf(`
resource "digitalocean_tag" "foo" {
  name = "web"
}

resource "digitalocean_droplet" "foo" {
  count              = 2
  image              = "ubuntu-18-04-x64"
  name               = "droplet-%d-${count.index}"
  region             = "nyc3"
  size               = "512mb"
  private_networking = true
  tags               = ["${digitalocean_tag.foo.id}"]
}

resource "digitalocean_loadbalancer" "foo" {
  name   = "loadbalancer-%d"
  region = "nyc3"

  forwarding_rule {
	entry_port     = 80
	entry_protocol = "http"

	target_port     = 80
	target_protocol = "http"
  }

  healthcheck {
	port     = 22
    protocol = "tcp"
  }

  droplet_tag = "${digitalocean_tag.foo.id}"
  depends_on  = ["digitalocean_droplet.foo"]
}

data "digitalocean_loadbalancer" "foobar" {
  name = "${digitalocean_loadbalancer.foo.name}"
}`, rInt, rInt)
}
