// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package infrastructure

import (
	"context"

	api "github.com/gardener/gardener-extension-provider-gcp/pkg/apis/gcp"
	infrainternal "github.com/gardener/gardener-extension-provider-gcp/pkg/internal/infrastructure"

	"github.com/gardener/gardener/extensions/pkg/controller/common"
	"github.com/gardener/gardener/extensions/pkg/controller/infrastructure"
	"github.com/gardener/gardener/extensions/pkg/terraformer"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type actuator struct {
	logger logr.Logger
	common.RESTConfigContext
	disableProjectedTokenMount bool
}

// NewActuator creates a new infrastructure.Actuator.
func NewActuator(disableProjectedTokenMount bool) infrastructure.Actuator {
	return &actuator{
		logger:                     log.Log.WithName("infrastructure-actuator"),
		disableProjectedTokenMount: disableProjectedTokenMount,
	}
}

func (a *actuator) updateProviderStatus(
	ctx context.Context,
	tf terraformer.Terraformer,
	infra *extensionsv1alpha1.Infrastructure,
	config *api.InfrastructureConfig,
) error {
	status, err := infrainternal.ComputeStatus(ctx, tf, config)
	if err != nil {
		return err
	}

	state, err := tf.GetRawState(ctx)
	if err != nil {
		return err
	}

	stateByte, err := state.Marshal()
	if err != nil {
		return err
	}

	patch := client.MergeFrom(infra.DeepCopy())
	infra.Status.ProviderStatus = &runtime.RawExtension{Object: status}
	infra.Status.State = &runtime.RawExtension{Raw: stateByte}
	return a.Client().Status().Patch(ctx, infra, patch)
}
