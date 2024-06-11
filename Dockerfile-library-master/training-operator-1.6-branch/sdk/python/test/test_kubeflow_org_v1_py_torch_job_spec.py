# coding: utf-8

"""
    Kubeflow Training SDK

    Python SDK for Kubeflow Training  # noqa: E501

    The version of the OpenAPI document: v1.6.0
    Generated by: https://openapi-generator.tech
"""


from __future__ import absolute_import

import unittest
import datetime

from kubeflow.training.models import *
from kubeflow.training.models.kubeflow_org_v1_py_torch_job_spec import KubeflowOrgV1PyTorchJobSpec  # noqa: E501
from kubeflow.training.rest import ApiException

class TestKubeflowOrgV1PyTorchJobSpec(unittest.TestCase):
    """KubeflowOrgV1PyTorchJobSpec unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def make_instance(self, include_optional):
        """Test KubeflowOrgV1PyTorchJobSpec
            include_option is a boolean, when False only required
            params are included, when True both required and
            optional params are included """
        # model = kubeflow.training.models.kubeflow_org_v1_py_torch_job_spec.KubeflowOrgV1PyTorchJobSpec()  # noqa: E501
        if include_optional :
            return KubeflowOrgV1PyTorchJobSpec(
                elastic_policy = kubeflow_org_v1_elastic_policy.KubeflowOrgV1ElasticPolicy(
                    max_replicas = 56, 
                    max_restarts = 56, 
                    metrics = [
                        None
                        ], 
                    min_replicas = 56, 
                    n_proc_per_node = 56, 
                    rdzv_backend = '0', 
                    rdzv_conf = [
                        kubeflow_org_v1_rdzv_conf.KubeflowOrgV1RDZVConf(
                            key = '0', 
                            value = '0', )
                        ], 
                    rdzv_host = '0', 
                    rdzv_id = '0', 
                    rdzv_port = 56, 
                    standalone = True, ), 
                pytorch_replica_specs = {
                    'key' : V1ReplicaSpec(
                        replicas = 56, 
                        restart_policy = '0', 
                        template = None, )
                    }, 
                run_policy = V1RunPolicy(
                    active_deadline_seconds = 56, 
                    backoff_limit = 56, 
                    clean_pod_policy = '0', 
                    scheduling_policy = V1SchedulingPolicy(
                        min_available = 56, 
                        min_resources = {
                            'key' : None
                            }, 
                        priority_class = '0', 
                        queue = '0', 
                        schedule_timeout_seconds = 56, ), 
                    ttl_seconds_after_finished = 56, )
            )
        else :
            return KubeflowOrgV1PyTorchJobSpec(
                pytorch_replica_specs = {
                    'key' : V1ReplicaSpec(
                        replicas = 56, 
                        restart_policy = '0', 
                        template = None, )
                    },
                run_policy = V1RunPolicy(
                    active_deadline_seconds = 56, 
                    backoff_limit = 56, 
                    clean_pod_policy = '0', 
                    scheduling_policy = V1SchedulingPolicy(
                        min_available = 56, 
                        min_resources = {
                            'key' : None
                            }, 
                        priority_class = '0', 
                        queue = '0', 
                        schedule_timeout_seconds = 56, ), 
                    ttl_seconds_after_finished = 56, ),
        )

    def testKubeflowOrgV1PyTorchJobSpec(self):
        """Test KubeflowOrgV1PyTorchJobSpec"""
        inst_req_only = self.make_instance(include_optional=False)
        inst_req_and_optional = self.make_instance(include_optional=True)


if __name__ == '__main__':
    unittest.main()
