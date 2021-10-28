package setting

import (
	"github.com/sirupsen/logrus"

	appsv1 "github.com/rancher/wrangler/pkg/generated/controllers/apps/v1"

	harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
)

// Handler updates the log level on setting changes
type Handler struct {
	dsClient appsv1.DaemonSetClient
	dsCache  appsv1.DaemonSetCache
}

func (h *Handler) LogLevelOnChanged(key string, setting *harvesterv1.Setting) (*harvesterv1.Setting, error) {
	if setting == nil || setting.DeletionTimestamp != nil || setting.Name != "log-level" || setting.Value == "" {
		return setting, nil
	}

	level, err := logrus.ParseLevel(setting.Value)
	if err != nil {
		return setting, err
	}

	logrus.Infof("set log level to %s", level)
	logrus.SetLevel(level)
	return setting, nil
}

func (h *Handler) AutoAddDiskPathsOnChanged(key string, setting *harvesterv1.Setting) (*harvesterv1.Setting, error) {
	if setting == nil || setting.DeletionTimestamp != nil || setting.Name != "auto-add-disk-paths" {
		return setting, nil
	}

	logrus.Debug("AudoAddDiskPathOnChanged")
	ds, err := h.dsCache.Get("harvester-system", "harvester-node-disk-manager")
	if err != nil {
		return setting, nil
	}

	dsCopy := ds.DeepCopy()

	for i, cont := range dsCopy.Spec.Template.Spec.Containers {
		if cont.Name == "harvester-node-disk-manager" {
			for j, env := range cont.Env {
				if env.Name == "NDM_AUTO_ADD_PATH" {
					logrus.Infof("update NDM_AUTO_ADD_PATH to '%s'", setting.Value)
					dsCopy.Spec.Template.Spec.Containers[i].Env[j].Value = setting.Value
				}
			}
		}
	}

	logrus.Info(dsCopy.Spec.Template.Spec.Containers)

	_, err = h.dsClient.Update(dsCopy)
	if err != nil {
		logrus.Errorf("unable to update NDM daemonset: %v", err)
		return setting, err
	}

	return setting, nil
}
