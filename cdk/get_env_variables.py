import os

ENVIRONMENT_VARIABLES = [
    'AWS_REGION',
    'ACCOUNT_ID',
    'SERVICE_NAME',
    'MD_REST_EFS_FOLDER_NAME',
    'MC_EMAIL_EFS_FOLDER_NAME',
    'IMAGE_TAG',
    'SERVICE_CPU',
    'SERVICE_MEMORY',
    'SERVICE_CONTAINER_PORT',
    'SERVICE_HOST_PORT',
    'OUTBOX_TABLE_NAME_PARAMETER_NAME',
    'MC_EML_EFS_ACCESS_POINT_ARN_PARAMETER_NAME',
    'MC_EML_EFS_ACCESS_POINT_ID_PARAMETER_NAME',
    'MC_EML_EFS_ID_PARAMETER_NAME',
    'REPOSITORY_NAME_PARAMETER_NAME',
    'MD_REST_EFS_ID_PARAMETER_NAME',
    'MD_REST_ACCESS_POINT_ID_PARAMETER_NAME',
    'MD_REST_ACCESS_POINT_ARN_PARAMETER_NAME',
    'TMP_TASK_DEFINITION_ARN_PARAMETER_NAME'
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
