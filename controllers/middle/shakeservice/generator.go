package shakeservice

import (
	"fmt"
	"strings"

	middlev1alpha1 "github.com/DevineLiu/redis-operator/apis/middle/v1alpha1"
	util2 "github.com/DevineLiu/redis-operator/controllers/util"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func generateRedisShakeService(rs *middlev1alpha1.RedisShake, labels map[string]string, ownrf []metav1.OwnerReference) *corev1.Service {
	name := util2.GetRedisShakeName(rs)
	namespace := rs.Namespace
	MetricTargetPort := intstr.FromInt(9320)
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ShakeRoleName, rs.Name))
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis-shake",
					Port:       9320,
					TargetPort: MetricTargetPort,
					Protocol:   "TCP",
				},
			},
		},
	}

}

func generateRedisShakeInitConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, ownrf []metav1.OwnerReference) *corev1.ConfigMap {
	name := util2.GetRedisShakeInitScriptName(rs)
	namespace := rs.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ShakeName, rs.Name))
	initScriptContent := `
cp  /config_tmp/%s /config/%s
sed -i s/{source_passwd}/${SOURCE_REDIS_PASSWORD}/g  /config/%s
sed -i s/{target_passwd}/${TARGET_REDIS_PASSWORD}/g  /config/%s
 	`
	ip := "127.0.0.1"
	port := "6379"
	if len(rs.Spec.TargetInfo.Address) > 0 {
		addr := rs.Spec.TargetInfo.Address[0]
		ip = strings.Split(addr, ":")[0]
		port = strings.Split(addr, ":")[1]

	}
	if rs.Spec.SourceInfo.Type == middlev1alpha1.Cluster && rs.Spec.SourceInfo.ClusterName != "" {
		ip = fmt.Sprintf("%s-0", rs.Spec.SourceInfo.ClusterName)
		port = "6379"
	}
	if rs.Spec.SourceInfo.Type == middlev1alpha1.Sentinel && rs.Spec.SourceInfo.ClusterName != "" {
		ip = fmt.Sprintf("rfr-%s-read-write", rs.Spec.SourceInfo.ClusterName)
		port = "6379"
	}

	extend_content := `
if [[ -z "${TARGET_REDIS_PASSWORD}" ]]; then
	
	redisShakeFlag=$(redis-cli -h %s -p %s  get redisShakeStopFlag|grep true)
	if [ "$redisShakeFlag" ] ; then
		echo "redis-shake block by flag"
		exit 1
	fi
	redisShakeFlag=$(redis-cli -c  -h %s -p %s  get redisShakeStopFlag|grep true)
	if [  "$redisShakeFlag" ] ; then
	echo "redis-shake block by flag"
		exit 1
	fi
else 
	redisShakeFlag=$(redis-cli -h %s -p %s -a ${TARGET_REDIS_PASSWORD} get redisShakeStopFlag|grep true)
	if [ "$redisShakeFlag" ] ; then
		echo "redis-shake block by flag"
		exit 1
	fi
	redisShakeFlag=$(redis-cli -c  -h %s -p %s -a ${TARGET_REDIS_PASSWORD} get redisShakeStopFlag|grep true)
	if [  "$redisShakeFlag" ] ; then
	echo "redis-shake block by flag"
		exit 1
	fi
fi
exit 0
`
	extend_content = fmt.Sprintf(extend_content, ip, port, ip, port, ip, port, ip, port)
	initScriptContent = fmt.Sprintf(initScriptContent, util2.ShakeConfigFileName, util2.ShakeConfigFileName, util2.ShakeConfigFileName, util2.ShakeConfigFileName)
	cmContent := initScriptContent + extend_content
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Data: map[string]string{
			"init.sh": cmContent,
		},
	}
}

func generateRedisShakeConfigMap(rs *middlev1alpha1.RedisShake, labels map[string]string, ownrf []metav1.OwnerReference) *corev1.ConfigMap {
	name := util2.GetRedisShakeName(rs)
	namespace := rs.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ShakeName, rs.Name))
	source_addr := ""
	target_addr := ""
	parallel := 32
	qbs := 200000
	key_exists := middlev1alpha1.NOne
	if rs.Spec.SourceInfo != nil {
		if len(rs.Spec.SourceInfo.Address) > 0 {
			source_addr = Slice2String(rs.Spec.SourceInfo.Address)
		}
	}

	if rs.Spec.SourceInfo.Type == middlev1alpha1.Cluster && rs.Spec.SourceInfo.ClusterName != "" {
		source_addr = fmt.Sprintf("master@%s-0:6379;%s-1:6379;%s-0:6379", rs.Spec.SourceInfo.ClusterName, rs.Spec.SourceInfo.ClusterName, rs.Spec.SourceInfo.ClusterName)
	}

	if rs.Spec.SourceInfo.Type == middlev1alpha1.Sentinel && rs.Spec.SourceInfo.ClusterName != "" {
		source_addr = fmt.Sprintf("mymaster@rfs-%s:26379", rs.Spec.SourceInfo.ClusterName)
	}

	if rs.Spec.TargetInfo.Type == middlev1alpha1.Cluster && rs.Spec.TargetInfo.ClusterName != "" {
		target_addr = fmt.Sprintf("%s-0:6379;%s-1:6379;%s-0:6379", rs.Spec.TargetInfo.ClusterName, rs.Spec.TargetInfo.ClusterName, rs.Spec.TargetInfo.ClusterName)
	}

	if rs.Spec.TargetInfo.Type == middlev1alpha1.Sentinel && rs.Spec.TargetInfo.ClusterName != "" {
		target_addr = fmt.Sprintf("mymaster@rfs-%s:26379", rs.Spec.TargetInfo.ClusterName)
	}

	if rs.Spec.TargetInfo != nil {
		if len(rs.Spec.TargetInfo.Address) > 0 {
			target_addr = Slice2String(rs.Spec.TargetInfo.Address)

		}
	}
	if rs.Spec.Parallel > 0 {
		parallel = int(rs.Spec.Parallel)
	}

	if rs.Spec.QBS > 0 {
		qbs = int(rs.Spec.QBS)
	}
	if rs.Spec.KeyExists != nil {
		key_exists = *rs.Spec.KeyExists
	}

	ShakeConfigMapContent := `conf.version = 1
id = redis-shake
log.file =
log.level = info
pid_path = 
system_profile = 9310
http_profile = 9320
parallel = %d
source.type = %s
source.address = %s
source.password_raw = {source_passwd}
source.auth_type = auth
source.tls_enable = %t
source.tls_skip_verify = %t
source.rdb.input =
source.rdb.parallel = 0
source.rdb.special_cloud = 
target.type = %s
target.address = %s
target.password_raw = {target_passwd}
target.db = -1
target.dbmap =
target.tls_enable = %t
target.tls_skip_verify = %t
target.rdb.output = local_dump
target.version =
fake_time =
key_exists = %s


filter.db.whitelist =
filter.db.blacklist =
filter.key.whitelist =
filter.key.blacklist =
filter.slot =
filter.command.whitelist =
filter.command.blacklist =
filter.lua = false

big_key_threshold = 524288000


metric = true
metric.print_log = false


sender.size = 104857600

sender.count = 4095
sender.delay_channel_size = 65535
keep_alive = 0

scan.key_number = 50
scan.special_cloud =
scan.key_file =

qps = %d
resume_from_break_point = %t

replace_hash_tag = false
`
	ConfigMapContent := fmt.Sprintf(ShakeConfigMapContent, parallel, rs.Spec.SourceInfo.Type, source_addr, rs.Spec.SourceInfo.TlsEnable, rs.Spec.SourceInfo.TlsSkipVerify,
		rs.Spec.TargetInfo.Type, target_addr, rs.Spec.TargetInfo.TlsEnable, rs.Spec.SourceInfo.TlsSkipVerify, key_exists, qbs, rs.Spec.ResumeFromBreakPoint,
	)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
		},
		Data: map[string]string{
			util2.ShakeConfigFileName: ConfigMapContent,
		},
	}

}

func Slice2String(slice []string) string {
	return strings.Join(slice, ";")
}

func generateRedisShakeDeployment(rs *middlev1alpha1.RedisShake, labels map[string]string, ownrf []metav1.OwnerReference) *v1.Deployment {
	name := util2.GetRedisShakeName(rs)
	namespace := rs.Namespace
	labels = util2.MergeMap(labels, generateSelectorLabels(util2.ShakeRoleName, rs.Name))
	ShakeCommand := generateShakeCommand(rs)
	deploy := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			Labels:          labels,
			OwnerReferences: ownrf,
			Annotations:     rs.Annotations,
		},
		Spec: v1.DeploymentSpec{
			Replicas: &rs.Spec.Replicas,
			Strategy: v1.DeploymentStrategy{
				Type: "RollingUpdate",
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: rs.Spec.PodAnnotations,
				},
				Spec: corev1.PodSpec{
					Affinity:         getAffinity(rs.Spec.Affinity, labels),
					Tolerations:      rs.Spec.Tolerations,
					NodeSelector:     rs.Spec.NodeSelector,
					SecurityContext:  getSecurityContext(rs.Spec.SecurityContext),
					ImagePullSecrets: rs.Spec.ImagePullSecrets,
					Containers: []corev1.Container{
						{
							Name:            "redis-shake",
							Image:           rs.Spec.Image,
							ImagePullPolicy: pullPolicy(rs.Spec.ImagePullPolicy),
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis-proxy",
									ContainerPort: 9320,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shake-config",
									MountPath: "/config_tmp",
								},
								{
									Name:      "shake-data",
									MountPath: "/config",
								},
							},
							Command:   ShakeCommand,
							Resources: rs.Spec.Resources,
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "shake-init",
							Image:           "build-harbor.alauda.cn/3rdparty/redis:5.0-alpine",
							ImagePullPolicy: pullPolicy(rs.Spec.ImagePullPolicy),
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "shake-config",
									MountPath: "/config_tmp",
								},
								{
									Name:      "shake-data",
									MountPath: "/config",
								},
								{
									Name:      "shake-init",
									MountPath: "/script",
								},
							},
							Command:   []string{"sh", "/script/init.sh"},
							Resources: rs.Spec.Resources,
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "shake-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: util2.GetRedisShakeName(rs),
									},
								},
							},
						},
						{
							Name: "shake-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "shake-init",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: util2.GetRedisShakeInitScriptName(rs),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if rs.Spec.SourceInfo.PasswordSecret != "" {
		deploy.Spec.Template.Spec.InitContainers[0].Env = append(deploy.Spec.Template.Spec.InitContainers[0].Env, corev1.EnvVar{
			Name: "SOURCE_REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rs.Spec.SourceInfo.PasswordSecret,
					},
					Key: "password",
				},
			},
		})
	}

	if rs.Spec.TargetInfo.PasswordSecret != "" {
		deploy.Spec.Template.Spec.InitContainers[0].Env = append(deploy.Spec.Template.Spec.InitContainers[0].Env, corev1.EnvVar{
			Name: "TARGET_REDIS_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: rs.Spec.TargetInfo.PasswordSecret,
					},
					Key: "password",
				},
			},
		})
	}

	return deploy

}

func generateShakeCommand(rs *middlev1alpha1.RedisShake) []string {
	return []string{
		"/redis-shake",
		"-conf",
		fmt.Sprintf("/config/%s", util2.ShakeConfigFileName),
		"-type",
		string(*rs.Spec.ModelType),
	}
}

func getAffinity(affinity *corev1.Affinity, labels map[string]string) *corev1.Affinity {
	if affinity != nil {
		return affinity
	}

	// Return a SOFT anti-affinity
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 100,
					PodAffinityTerm: corev1.PodAffinityTerm{
						TopologyKey: util2.HostnameTopologyKey,
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: labels,
						},
					},
				},
			},
		},
	}
}

func getSecurityContext(secctx *corev1.PodSecurityContext) *corev1.PodSecurityContext {
	if secctx != nil {
		return secctx
	}

	return nil
}

func pullPolicy(specPolicy corev1.PullPolicy) corev1.PullPolicy {
	if specPolicy == "" {
		return corev1.PullAlways
	}
	return specPolicy
}
