#!/usr/bin/env python3
from aws_cdk import App, Environment

from get_env_variables import GetEnvVariables
from task_definition_stack import TaskDefinitionStack

if __name__ == "__main__":
    app = App()

    selected_environment = app.node.try_get_context('environment')

    env_parameters = GetEnvVariables(selected_environment).env_dict

    environment = Environment(account=env_parameters['ACCOUNT_ID'], region=env_parameters['AWS_REGION'])

    TaskDefinitionStack(
        app,
        f"{env_parameters['SELECTED_ENVIRONMENT']}-multicarrier-email-api-task-definition-stack",
        env_parameters=env_parameters,
        env=environment
    )

    app.synth()