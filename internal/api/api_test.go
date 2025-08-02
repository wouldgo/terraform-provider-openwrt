package api

import (
	"context"
	"os"
	"testing"
)

var (
	openWRTRemoteEnv,
	openWRTRemoteEnvSet = os.LookupEnv("OPENWRT_REMOTE")

	openWRTUserEnv,
	openWRTUserEnvSet = os.LookupEnv("OPENWRT_USER")

	openWRTPasswordEnv,
	openWRTPasswordEnvSet = os.LookupEnv("OPENWRT_PASSWORD")
)

func TestLuciApis(t *testing.T) {
	if !openWRTRemoteEnvSet ||
		!openWRTUserEnvSet ||
		!openWRTPasswordEnvSet {
		panic("please specify OPENWRT_URL, OPENWRT_ROOT_USER and OPENWRT_ROOT_PASSWORD env vars to run the tests")
	}

	testCases := []struct {
		desc string
		test func(*testing.T, Client)
	}{
		{
			desc: "uci GetAll no sections",
			test: func(t *testing.T, c Client) {
				result, err := c.GetAll(context.Background())
				if err == nil {
					t.Error(err)
				}

				if result != nil {
					t.Error("result expected to be nil")
				}

				if err.Error() != "no sections specified" {
					t.Error("expected \"no sections specified\" returned: %w", err)
				}
			},
		},
		{
			desc: "uci GetSystem",
			test: func(t *testing.T, c Client) {
				system, err := c.GetSystem(context.Background())
				if err != nil {
					t.Error(err)
				}

				if system == nil {
					t.Error("system must be present")
				}
			},
		},

		// {
		// 	desc: "write file",
		// 	test: func(t *testing.T, c Client) error {
		// 		err := c.Writefile("/tmp/test.txt", []byte("hello"))
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "get package info",
		// 	test: func(t *testing.T, c Client) error {
		// 		response, err := c.CheckPackage("curl")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		t.Logf("check package: %+v", response)
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "update packages ref",
		// 	test: func(t *testing.T, c Client) error {
		// 		err := c.UpdatePackages()
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "install packages",
		// 	test: func(t *testing.T, c Client) error {
		// 		err := c.InstallPackage("curl")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "remove packages",
		// 	test: func(t *testing.T, c Client) error {
		// 		err := c.RemovePackages("curl")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "get all",
		// 	test: func(t *testing.T, client Client) error {
		// 		raw, err := client.UCIGetAll("system")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		fmt.Println(raw)
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "get system",
		// 	test: func(t *testing.T, c Client) error {
		// 		raw, err := c.UCIGetSystem()
		// 		if err != nil {
		// 			return err
		// 		}
		// 		fmt.Println(raw)
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "get system ntp",
		// 	test: func(t *testing.T, c Client) error {
		// 		raw, err := c.UCIGetAll("system", "ntp")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		fmt.Println(raw)
		// 		return nil
		// 	},
		// },
		// {
		// 	desc: "get specific file",
		// 	test: func(t *testing.T, c Client) error {
		// 		raw, err := c.ReadFile("/etc/config/system")
		// 		if err != nil {
		// 			return err
		// 		}
		// 		fmt.Println(raw)
		// 		return nil
		// 	},
		// },
	}

	t.Run("passing empty string as openwrt remote", func(t *testing.T) {
		c, err := NewClient("")
		if err == nil {
			t.Error("error must be present")
		}

		if c != nil {
			t.Error("client must be nil")
		}
	})

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			c, err := NewClient(openWRTRemoteEnv)
			if err != nil {
				t.Fatal(err)
			}
			err = c.Auth(context.Background(), openWRTUserEnv, openWRTPasswordEnv)
			if err != nil {
				t.Fatal(err)
			}
			tC.test(t, c)
		})
	}
}
