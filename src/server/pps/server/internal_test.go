package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestAPIServer_validatePipelineRequest(t *testing.T) {
	var (
		a       = new(apiServer)
		request = &pps.CreatePipelineRequest{
			Pipeline: &pps.Pipeline{
				Project: &pfs.Project{
					Name: "0123456789ABCDEF",
				},
				Name: "0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789ABCDEF",
			},
		}
	)
	err := a.validatePipelineRequest(request)
	require.YesError(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("is %d characters longer than the %d max", len(request.Pipeline.Name)-maxPipelineNameLength, maxPipelineNameLength))

	request.Pipeline.Name = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	k8sName := ppsutil.PipelineRcName(&pps.PipelineInfo{Pipeline: &pps.Pipeline{Project: &pfs.Project{Name: request.Pipeline.GetProject().GetName()}, Name: request.Pipeline.Name}, Version: 99})
	err = a.validatePipelineRequest(request)
	require.YesError(t, err)
	require.ErrorContains(t, err, fmt.Sprintf("is %d characters longer than the %d max", len(k8sName)-dnsLabelLimit, dnsLabelLimit))
}

func TestNewMessageFilterFunc(t *testing.T) {
	ctx := pctx.TestContext(t)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	filter, err := newMessageFilterFunc(`{ source: ., output: "" } | until(.source == "quux"; {"foo": "bar"}) | .output`, nil)
	require.NoError(t, err)
	go func() {
		_, err = filter(ctx, &pps.JobInfo{})
		require.YesError(t, err)
	}()
	<-ctx.Done()
	require.Equal(t, context.DeadlineExceeded, ctx.Err())
}

func BenchmarkMakeEffectiveSpec(b *testing.B) {
	var (
		clusterDefaults = map[string]string{
			"empty defaults":                      `{}`,
			"empty default createPipelineRequest": `{"createPipelineRequest":{}}`,
			"default autoscaling":                 `{"createPipelineRequest":{"autoscaling": true}}`,
			"default datum_tries":                 `{"create_pipeline_request":{"datum_tries": 6}}`,
			"default autoscaling and datumTries":  `{"createPipelineRequest":{"autoscaling": false, "datumTries": 6}}`,
			"default resource requests and limits": `{"createPipelineRequest":{
				"resource_requests": {
					"cpu": "1",
					"disk": "187Mi"
				},
				"resource_limits": {
					"cpu": "1",
					"disk": "187Mi"
				}
			}}`,
		}
		projectDefaults = map[string]string{
			"empty defaults":                      `{}`,
			"empty default createPipelineRequest": `{"createPipelineRequest":{}}`,
			"default autoscaling":                 `{"createPipelineRequest":{"autoscaling": true}}`,
			"default datum_tries":                 `{"create_pipeline_request":{"datum_tries": 7}}`,
			"default autoscaling and datumTries":  `{"createPipelineRequest":{"autoscaling": false, "datumTries": 4}}`,
			"default resource requests and limits": `{"createPipelineRequest":{
				"resource_requests": {
					"cpu": "1",
					"disk": "186Mi"
				},
				"resource_limits": {
					"cpu": "1",
					"disk": "188Mi"
				}
			}}`,
		}
		specs = map[string]string{
			"empty spec": `{}`,
			"simple spec": `{
    "pipeline": {
	"name": "simple"
    },
    "input": {
	"pfs": {
	    "glob": "/*",
	    "repo": "input"
	}
    },
    "parallelism_spec": {
	"constant": 1
    },
    "transform": {
	"cmd": [ "/bin/bash" ],
	"stdin": [
	    "cp /pfs/input/* /pfs/out"
	]
    },
    "resourceRequests": {
	"cpu": null,
	"disk": "256Mi"
    }
}`,
			"complex spec": `{
  "pipeline": {
    "name": "yz63tb0lsrttr1o5vi6hnv"
  },
  "transform": {
    "image": "v8kikjkksj6wiowylr43s307i0nnqt1v1xf4tqor9",
    "cmd": [
      "/bin/bash"
    ],
    "stdin": [
      "#!/bin/bash",
      "g865rbaliyxugol84xtysyo6kaavagfzw1iwpfaxf6tjeq8"
    ],
    "env": {
      "LOG_LEVEL": "DEBUG",
      "OUT_PATH": "/pfs/out",
      "KOYV01aAVT3GLGD3ZP": "ptxtdkse4c46zdzdmn6s0s7nbgjur0thywre7gf2fc5ztqk23f9sh",
      "QOR53AQIPHQaJUAWvSVDUG": "k",
      "LMT7212QQ4AKDGF": "j",
      "7EC7oVOVEYdX7Q18": "v",
      "3D076CSLxAZERS": "6",
      "109W7NXS4DI0BaLG4IY": "3",
      "D1ZNfIVD529TNKC": "x",
      "6M36JUUC1UG3KPNeJYB6pD3QIH": "b",
      "NNWSNNZFzPUJGdJOWA8": "p",
      "SN7ETONUtCCRI2cFNFST": "q",
      "DT2C9NHKg8XMlFT8GS": "r",
      "N0DYPR1Gt4Z0RE": "h",
      "PBEPGcGDGQp46RCuCH8GL": "b"
    }
  },
  "parallelism_spec": {
    "constant": "16"
  },
  "input": {
    "cross": [
      {
        "pfs": {
          "name": "PEF2CwFUGBlSC9F",
          "repo": "466nlvu6pmb",
          "glob": "nk3fqr6"
        }
      },
      {
        "union": [
          {
            "pfs": {
              "name": "YLBDiZ1SU",
              "repo": "lwkiwc4t6q2gfsj2xk3vrmdbm7to8",
              "glob": "/0m1ytf/(*)/(*)/(*)",
              "joinOn": "$1/$2/$3"
            }
          },
          {
            "join": [
              {
                "pfs": {
                  "name": "7QGNJ3REq7OW6",
                  "repo": "f3j93anf3bcyjkoidv12qlq7l8s8fpb0j8okz8x",
                  "glob": "/fhgptj/(*)/(*)/(*)/(MU5YH7n7om78|LEHR3K01o9va|I77WHKzlepzg)",
                  "joinOn": "$1/$2/$3"
                }
              },
              {
                "pfs": {
                  "name": "LQIUnZQ24WFQvNJKB",
                  "repo": "dzb49cse84hat6xbjqgiudxm5hw9gp",
                  "glob": "/(*)/(*)/(*)",
                  "joinOn": "$1/$2/$3"
                }
              }
            ]
          }
        ]
      }
    ]
  },
  "autoscaling": true
}`,
			"complex spec 2": `{
  "pipeline": {
    "name": "iki30apxti994kpi1wnsko"
  },
  "transform": {
    "image": "plmxrfvb9h4nsfjvc3nj56o3tu3kjtarxlzqpz4mb",
    "cmd": [
      "/bin/bash"
    ],
    "stdin": [
      "#!/bin/bash",
      "ffqu4wcxeh39iwa4s6mls2wx6uu6f3taappoxqs90ptdu57"
    ],
    "env": {
      "LOG_LEVEL": "DEBUG",
      "OUT_PATH": "/pfs/out",
      "DJOLV20H89Z3INX6SF": "gxjewzd40kciifcjw4xi4lg97kuxji37rtme0e092o169sveco7ul",
      "JMVVi64RAT0lKEQBr9GI08": "q",
      "WTHVpL1VOo6JCBQ": "s",
      "6GAX2AEZ8N54WPMQ": "u",
      "8ZA60EP0yCCXEF": "y",
      "3XKNeQ0G0GCGJk99RVE": "c",
      "MJF2929EPwUP184": "5",
      "C5M3WYLY9E5I6QBxSJENl463Q1": "i",
      "3CW6X6WWjXJL2u2VWIH": "t",
      "JAI0684D8XP1ELeSE43N": "v",
      "W6MDSQEDkKZXw6EGUY": "r",
      "PF0KP4PV5XAG2X": "9",
      "A2NP6bP6M1hA4RQy00Y13": "f"
    }
  },
  "parallelism_spec": {
    "constant": "4"
  },
  "input": {
    "cross": [
      {
        "pfs": {
          "name": "H56IMlYJKJr8K5U",
          "repo": "xte9lujxlfz",
          "glob": "/ya9d0d"
        }
      },
      {
        "union": [
          {
            "pfs": {
              "name": "ETO2oHU37",
              "repo": "txz932n7gxao0szqmymep7vb56j44",
              "glob": "/ggigb9/(*)/(*)/(*)",
              "joinOn": "$1/$2/$3"
            }
          },
          {
            "join": [
              {
                "pfs": {
                  "name": "3KWHBGKKl42D6",
                  "repo": "kzfzxzkjv02pwimihxxl5y0ihwa6gvwqcpt5of9",
                  "glob": "/9sgzyc/(*)/(*)/(*)/(TDH39N8n6f98|0YES2Yldixlj|IW4IWFlwof2r)",
                  "joinOn": "$1/$2/$3"
                }
              },
              {
                "pfs": {
                  "name": "LXKVmRYBHBF6pLLG9",
                  "repo": "vems13rvb1zzj97xg57mef57dt9655",
                  "glob": "/(*)/(*)/(*)",
                  "joinOn": "$1/$2/$3"
                }
              }
            ]
          }
        ]
      }
    ]
  },
  "autoscaling": true
}`,
		}
	)
	// merge each spec into each default
	for name, clusterDefaults := range clusterDefaults {
		b.Run(name, func(b *testing.B) {
			for name, projectDefaults := range projectDefaults {
				b.Run(name, func(b *testing.B) {
					for name, spec := range specs {
						b.Run(name, func(b *testing.B) {
							for i := 0; i < b.N; i++ {
								if _, _, err := makeEffectiveSpec(clusterDefaults, projectDefaults, spec); err != nil {
									b.Error(err)
								}

							}
						})
					}
				})
			}
		})
	}
}
