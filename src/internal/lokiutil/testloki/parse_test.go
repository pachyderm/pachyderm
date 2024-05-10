package testloki

import (
	"fmt"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestParseLabels(t *testing.T) {
	testData := []struct {
		input string
		want  map[string]string
	}{
		{
			input: `&map[app:console container:console filename:/var/log/pods/default_console-86d4769794-jwtk2_def035a4-378e-4971-a2a4-80090ee7f4cc/console/0.log job:default/console namespace:default node_name:pach-control-plane pod:console-86d4769794-jwtk2 pod_template_hash:86d4769794 stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":               "console",
				"container":         "console",
				"filename":          "/var/log/pods/default_console-86d4769794-jwtk2_def035a4-378e-4971-a2a4-80090ee7f4cc/console/0.log",
				"job":               "default/console",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "console-86d4769794-jwtk2",
				"pod_template_hash": "86d4769794",
				"stream":            "stderr",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:console container:console filename:/var/log/pods/default_console-86d4769794-jwtk2_def035a4-378e-4971-a2a4-80090ee7f4cc/console/0.log job:default/console namespace:default node_name:pach-control-plane pod:console-86d4769794-jwtk2 pod_template_hash:86d4769794 stream:stdout suite:pachyderm]`,
			want: map[string]string{
				"app":               "console",
				"container":         "console",
				"filename":          "/var/log/pods/default_console-86d4769794-jwtk2_def035a4-378e-4971-a2a4-80090ee7f4cc/console/0.log",
				"job":               "default/console",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "console-86d4769794-jwtk2",
				"pod_template_hash": "86d4769794",
				"stream":            "stdout",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:etcd apps_kubernetes_io_pod_index:0 container:etcd controller_revision_hash:etcd-5c748dc64f filename:/var/log/pods/default_etcd-0_d80ed43c-7439-4a97-a905-7e035e34d683/etcd/0.log job:default/etcd namespace:default node_name:pach-control-plane pod:etcd-0 statefulset_kubernetes_io_pod_name:etcd-0 stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":                                "etcd",
				"apps_kubernetes_io_pod_index":       "0",
				"container":                          "etcd",
				"controller_revision_hash":           "etcd-5c748dc64f",
				"filename":                           "/var/log/pods/default_etcd-0_d80ed43c-7439-4a97-a905-7e035e34d683/etcd/0.log",
				"job":                                "default/etcd",
				"namespace":                          "default",
				"node_name":                          "pach-control-plane",
				"pod":                                "etcd-0",
				"statefulset_kubernetes_io_pod_name": "etcd-0",
				"stream":                             "stderr",
				"suite":                              "pachyderm",
			},
		},
		{
			input: `&map[app:pachd container:pachd filename:/var/log/pods/default_pachd-59bbd54f97-vxb8d_7a3cbbfb-f6ec-44d3-80b5-8fea99daba57/pachd/0.log job:default/pachd namespace:default node_name:pach-control-plane pod:pachd-59bbd54f97-vxb8d pod_template_hash:59bbd54f97 stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":               "pachd",
				"container":         "pachd",
				"filename":          "/var/log/pods/default_pachd-59bbd54f97-vxb8d_7a3cbbfb-f6ec-44d3-80b5-8fea99daba57/pachd/0.log",
				"job":               "default/pachd",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "pachd-59bbd54f97-vxb8d",
				"pod_template_hash": "59bbd54f97",
				"stream":            "stderr",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:pachyderm-kube-event-tail container:kube-event-tail filename:/var/log/pods/default_pachyderm-kube-event-tail-7d7764b7c6-dlqgb_5578a06f-b9fd-4705-b06a-00af2a31bcd7/kube-event-tail/0.log job:default/pachyderm-kube-event-tail namespace:default node_name:pach-control-plane pod:pachyderm-kube-event-tail-7d7764b7c6-dlqgb pod_template_hash:7d7764b7c6 stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":               "pachyderm-kube-event-tail",
				"container":         "kube-event-tail",
				"filename":          "/var/log/pods/default_pachyderm-kube-event-tail-7d7764b7c6-dlqgb_5578a06f-b9fd-4705-b06a-00af2a31bcd7/kube-event-tail/0.log",
				"job":               "default/pachyderm-kube-event-tail",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "pachyderm-kube-event-tail-7d7764b7c6-dlqgb",
				"pod_template_hash": "7d7764b7c6",
				"stream":            "stderr",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:pachyderm-proxy container:envoy filename:/var/log/pods/default_pachyderm-proxy-5585cd764c-64hgc_4faadc59-5aa4-4af3-89b8-3ce73e7ac298/envoy/0.log job:default/pachyderm-proxy namespace:default node_name:pach-control-plane pod:pachyderm-proxy-5585cd764c-64hgc pod_template_hash:5585cd764c stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":               "pachyderm-proxy",
				"container":         "envoy",
				"filename":          "/var/log/pods/default_pachyderm-proxy-5585cd764c-64hgc_4faadc59-5aa4-4af3-89b8-3ce73e7ac298/envoy/0.log",
				"job":               "default/pachyderm-proxy",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "pachyderm-proxy-5585cd764c-64hgc",
				"pod_template_hash": "5585cd764c",
				"stream":            "stderr",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:pachyderm-proxy container:envoy filename:/var/log/pods/default_pachyderm-proxy-5585cd764c-64hgc_4faadc59-5aa4-4af3-89b8-3ce73e7ac298/envoy/0.log job:default/pachyderm-proxy namespace:default node_name:pach-control-plane pod:pachyderm-proxy-5585cd764c-64hgc pod_template_hash:5585cd764c stream:stdout suite:pachyderm]`,
			want: map[string]string{
				"app":               "pachyderm-proxy",
				"container":         "envoy",
				"filename":          "/var/log/pods/default_pachyderm-proxy-5585cd764c-64hgc_4faadc59-5aa4-4af3-89b8-3ce73e7ac298/envoy/0.log",
				"job":               "default/pachyderm-proxy",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "pachyderm-proxy-5585cd764c-64hgc",
				"pod_template_hash": "5585cd764c",
				"stream":            "stdout",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:pg-bouncer container:pg-bouncer filename:/var/log/pods/default_pg-bouncer-7f45dc7865-cmrt7_acf0209c-33ea-4c48-931e-af68ac77e294/pg-bouncer/0.log job:default/pg-bouncer namespace:default node_name:pach-control-plane pod:pg-bouncer-7f45dc7865-cmrt7 pod_template_hash:7f45dc7865 stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":               "pg-bouncer",
				"container":         "pg-bouncer",
				"filename":          "/var/log/pods/default_pg-bouncer-7f45dc7865-cmrt7_acf0209c-33ea-4c48-931e-af68ac77e294/pg-bouncer/0.log",
				"job":               "default/pg-bouncer",
				"namespace":         "default",
				"node_name":         "pach-control-plane",
				"pod":               "pg-bouncer-7f45dc7865-cmrt7",
				"pod_template_hash": "7f45dc7865",
				"stream":            "stderr",
				"suite":             "pachyderm",
			},
		},
		{
			input: `&map[app:pipeline component:worker container:storage filename:/var/log/pods/default_default-edges-v1-292m8_2e2a393d-0202-4c39-b5f7-fcc312625369/storage/0.log job:default/pipeline namespace:default node_name:pach-control-plane pipelineName:edges pipelineProject:default pipelineVersion:1 pod:default-edges-v1-292m8 stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":             "pipeline",
				"component":       "worker",
				"container":       "storage",
				"filename":        "/var/log/pods/default_default-edges-v1-292m8_2e2a393d-0202-4c39-b5f7-fcc312625369/storage/0.log",
				"job":             "default/pipeline",
				"namespace":       "default",
				"node_name":       "pach-control-plane",
				"pipelineName":    "edges",
				"pipelineProject": "default",
				"pipelineVersion": "1",
				"pod":             "default-edges-v1-292m8",
				"stream":          "stderr",
				"suite":           "pachyderm",
			},
		},
		{
			input: `&map[app:pipeline component:worker container:storage filename:/var/log/pods/default_default-edges-v2-9f22j_7c25ca2b-89df-4906-90c0-89ffe32f56ec/storage/0.log job:default/pipeline namespace:default node_name:pach-control-plane pipelineName:edges pipelineProject:default pipelineVersion:2 pod:default-edges-v2-9f22j stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":             "pipeline",
				"component":       "worker",
				"container":       "storage",
				"filename":        "/var/log/pods/default_default-edges-v2-9f22j_7c25ca2b-89df-4906-90c0-89ffe32f56ec/storage/0.log",
				"job":             "default/pipeline",
				"namespace":       "default",
				"node_name":       "pach-control-plane",
				"pipelineName":    "edges",
				"pipelineProject": "default",
				"pipelineVersion": "2",
				"pod":             "default-edges-v2-9f22j",
				"stream":          "stderr",
				"suite":           "pachyderm",
			},
		},
		{
			input: `&map[app:pipeline component:worker container:storage filename:/var/log/pods/default_default-montage-v1-57sqv_03ae73e8-7041-4215-82c1-65b0bddf1126/storage/0.log job:default/pipeline namespace:default node_name:pach-control-plane pipelineName:montage pipelineProject:default pipelineVersion:1 pod:default-montage-v1-57sqv stream:stderr suite:pachyderm]`,
			want: map[string]string{
				"app":             "pipeline",
				"component":       "worker",
				"container":       "storage",
				"filename":        "/var/log/pods/default_default-montage-v1-57sqv_03ae73e8-7041-4215-82c1-65b0bddf1126/storage/0.log",
				"job":             "default/pipeline",
				"namespace":       "default",
				"node_name":       "pach-control-plane",
				"pipelineName":    "montage",
				"pipelineProject": "default",
				"pipelineVersion": "1",
				"pod":             "default-montage-v1-57sqv",
				"stream":          "stderr",
				"suite":           "pachyderm",
			},
		},
		{
			input: `&map[app:pipeline component:worker container:user filename:/var/log/pods/default_default-edges-v2-9f22j_7c25ca2b-89df-4906-90c0-89ffe32f56ec/user/0.log job:default/pipeline namespace:default node_name:pach-control-plane pipelineName:edges pipelineProject:default pipelineVersion:2 pod:default-edges-v2-9f22j stream:stdout suite:pachyderm]`,
			want: map[string]string{
				"app":             "pipeline",
				"component":       "worker",
				"container":       "user",
				"filename":        "/var/log/pods/default_default-edges-v2-9f22j_7c25ca2b-89df-4906-90c0-89ffe32f56ec/user/0.log",
				"job":             "default/pipeline",
				"namespace":       "default",
				"node_name":       "pach-control-plane",
				"pipelineName":    "edges",
				"pipelineProject": "default",
				"pipelineVersion": "2",
				"pod":             "default-edges-v2-9f22j",
				"stream":          "stdout",
				"suite":           "pachyderm",
			},
		},
		{
			input: `&map[app:pipeline component:worker container:user filename:/var/log/pods/default_default-montage-v1-57sqv_03ae73e8-7041-4215-82c1-65b0bddf1126/user/0.log job:default/pipeline namespace:default node_name:pach-control-plane pipelineName:montage pipelineProject:default pipelineVersion:1 pod:default-montage-v1-57sqv stream:stdout suite:pachyderm]`,
			want: map[string]string{
				"app":             "pipeline",
				"component":       "worker",
				"container":       "user",
				"filename":        "/var/log/pods/default_default-montage-v1-57sqv_03ae73e8-7041-4215-82c1-65b0bddf1126/user/0.log",
				"job":             "default/pipeline",
				"namespace":       "default",
				"node_name":       "pach-control-plane",
				"pipelineName":    "montage",
				"pipelineProject": "default",
				"pipelineVersion": "1",
				"pod":             "default-montage-v1-57sqv",
				"stream":          "stdout",
				"suite":           "pachyderm",
			},
		},
	}

	for i, test := range testData {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			got := parseLabels(test.input)
			require.NoDiff(t, test.want, got, nil)
		})
	}
}
