package blueprint

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/api/v1"
	v1alpha1app "k8s.io/client-go/pkg/apis/apps/v1alpha1"
	v2alpha1batch "k8s.io/client-go/pkg/apis/batch/v2alpha1"
	v1beta1ext "k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/pkg/util/intstr"
)

var (
	// TODO: Make this configurable
	DockerReg    = "gcr.io"
	Account      = "testacct"
	KubeFileType = "yaml"
)

func deployKubeFile(file string, defs []interface{}) error {

	btys, err := marshalJSON(defs)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(file, btys, 0644); err != nil {
		return fmt.Errorf("unable to write file '%s': %s", file, err)
	}

	return nil
}

func marshalJSON(stuff []interface{}) ([]byte, error) {
	btys := make([][]byte, len(stuff))
	for i, s := range stuff {
		//b, err := json.MarshalIndent(s, "", "  ")
		b, err := yaml.Marshal(s)
		if err != nil {
			return nil, err
		}
		btys[i] = b
	}

	return bytes.Join(btys, []byte("---\n")), nil
}

func kubeScheduledJob(name, schedule string) *v2alpha1batch.ScheduledJob {
	// TODO: how to get version?
	version := 1

	return &v2alpha1batch.ScheduledJob{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: v2alpha1batch.SchemeGroupVersion.String(),
			Kind:       "ScheduledJob",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v2alpha1batch.ScheduledJobSpec{
			Schedule:          schedule,
			ConcurrencyPolicy: v2alpha1batch.ForbidConcurrent,
			JobTemplate: v2alpha1batch.JobTemplateSpec{
				Spec: v2alpha1batch.JobSpec{
					Template: v1.PodTemplateSpec{
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  name + "-cron",
									Image: image(name, version),
								},
							},
							RestartPolicy: v1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}
}

func kubeService(idt Identifier, port int32) *v1.Service {
	return &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: idt.Host(),
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeLoadBalancer,
			Ports: []v1.ServicePort{
				{
					Port: port,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: port,
					},
					Protocol: v1.ProtocolTCP,
					// Name:     "p",
				},
			},
			Selector: map[string]string{
				"name": idt.Name,
				"kind": idt.Kind,
			},
		},
	}
}

func kubeDeploymentAPI(name string) *v1beta1ext.Deployment {

	// TODO: How to determine version number?
	version := 1

	return &v1beta1ext.Deployment{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: v1beta1ext.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1ext.DeploymentSpec{
			Selector: &unversioned.LabelSelector{
				MatchLabels: map[string]string{
					"name": name,
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"name": name,
						"kind": "api",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:    name + "-grpcd",
							Image:   image(name, version),
							Command: []string{"grpcd"},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "PORT",
									Value: "50051",
								},
							},
						},
						{
							Name:    name + "-httpd",
							Image:   image(name, version),
							Command: []string{"httpd"},
							Ports: []v1.ContainerPort{
								{
									ContainerPort: 8080,
								},
							},
							Env: []v1.EnvVar{
								{
									Name:  "PORT",
									Value: "8080",
								},
							},
						},
					},
				},
			},
		},
	}
}

func mustKubeStatefulSetDB(name, typ string) *v1alpha1app.StatefulSet {

	vol := "db-storage"

	var mnt, img string
	var envs []v1.EnvVar
	var port int32
	switch typ {
	case DBTypeMySQL:
		img = "mysql:5.7"
		mnt = "/var/lib/mysql"
		port = 3306
		envs = []v1.EnvVar{
			{
				Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
				Value: "yes",
			},
		}
	default:
		// This should never happen
		panic("unrecognized db-type")
	}

	return &v1alpha1app.StatefulSet{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: v1alpha1app.SchemeGroupVersion.String(),
			Kind:       "PetSet", //"StatefulSet",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1app.StatefulSetSpec{
			ServiceName: name,
			Replicas:    int32p(1),
			Template: v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"db":   typ,
						"kind": "db",
						"name": name,
					},
					Annotations: map[string]string{
						"pod.alpha.kubernetes.io/initialized": "true",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  name + "-db",
							Image: img,
							Ports: []v1.ContainerPort{
								{
									ContainerPort: port,
								},
							},
							Env: envs,
							VolumeMounts: []v1.VolumeMount{
								{
									Name:      vol,
									MountPath: mnt,
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: v1.ObjectMeta{
						Name: vol,
						Annotations: map[string]string{
							"volume.alpha.kubernetes.io/storage-class": "anything",
						},
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							v1.ReadWriteOnce,
						},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
	}
	return nil
}

func int32p(x int32) *int32 {
	return &x
}

func image(name string, v int) string {
	return fmt.Sprintf("%s/%s/%s:%v",
		DockerReg,
		Account,
		name,
		v,
	)
}
