import os

ENVIRONMENT_VARIABLES = [
    'AWS_REGION',
    'ACCOUNT_ID',
    'SERVICE_NAME',
    # 'CLUSTER_NAME',
    # 'VPC_ID',
    # 'MD_REST_EFS_ID',
    'MD_REST_EFS_FOLDER_NAME',
    # 'MD_REST_EFS_SECURITY_GROUP_ID',
    'MC_EMAIL_EFS_FOLDER_NAME',
    # 'MD_DOMAIN',
    # 'ALB_RULE_PRIORITY',
    # 'IMAGE_TAG',
    'SERVICE_CPU',
    'SERVICE_MEMORY',
    # 'SERVICE_DESIRED_COUNT',
    # 'SERVICE_MAX_COUNT',
    'SERVICE_CONTAINER_PORT',
    'SERVICE_HOST_PORT',
    # 'SSM_INTERNAL_ALB_ARN'
]

class GetEnvVariables:
    def __init__(
            self,
            selected_environment: str
    ) -> None:

        self.env_dict = {
            'SELECTED_ENVIRONMENT': selected_environment,
        }

        for i in ENVIRONMENT_VARIABLES:

            if i == 'IMAGE_TAG':
                self.env_dict[i] = 'latest'
                continue

            self.env_dict[i] = os.environ[i]
