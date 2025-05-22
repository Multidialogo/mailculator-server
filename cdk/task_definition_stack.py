from aws_cdk import (
    aws_ecs as ecs,
    aws_logs as logs,
    aws_iam as iam,
    aws_ec2 as ec2,
    aws_efs as efs,
    aws_dynamodb as dynamodb,
    aws_ssm as ssm,
    aws_elasticloadbalancingv2 as elbv2,
    aws_route53 as route53,
    aws_route53_targets as route53_targets,
    aws_ecr as ecr,
    aws_applicationautoscaling as applicationautoscaling,
    Stack,
    RemovalPolicy,
    Duration,
    TimeZone
)
from constructs import Construct

MD_REST_VOLUME_NAME = 'rest-volume'
MC_VOLUME_NAME = 'mc-volume'

MULTICARRIER_EMAIL_ID = 'multicarrier-email'

class TaskDefinitionStack(Stack):
    def __init__(
            self,
            scope: Construct,
            id: str,
            env_parameters: dict,
            **kwargs
    ) -> None:
        super().__init__(
            scope,
            id,
            **kwargs
        )

        service_name = env_parameters['SERVICE_NAME']
        selected_environment = env_parameters['SELECTED_ENVIRONMENT']
        # md_rest_efs_id = env_parameters['MD_REST_EFS_ID']
        md_rest_efs_folder_name = env_parameters['MD_REST_EFS_FOLDER_NAME']
        # md_rest_efs_security_group_id = env_parameters['MD_REST_EFS_SECURITY_GROUP_ID']
        mc_email_efs_folder_name = env_parameters['MC_EMAIL_EFS_FOLDER_NAME']
        image_tag = env_parameters['IMAGE_TAG']
        # vpc_id = env_parameters['VPC_ID']
        # cluster_name = env_parameters['CLUSTER_NAME']
        # md_domain = env_parameters['MD_DOMAIN']
        # alb_rule_priority = env_parameters['ALB_RULE_PRIORITY']
        service_cpu = env_parameters['SERVICE_CPU']
        service_memory = env_parameters['SERVICE_MEMORY']
        # service_desired_count = env_parameters['SERVICE_DESIRED_COUNT']
        # service_max_count = env_parameters['SERVICE_MAX_COUNT']
        service_container_port = env_parameters['SERVICE_CONTAINER_PORT']
        service_host_port = env_parameters['SERVICE_HOST_PORT']


        # cfn_efs_access_point = efs.CfnAccessPoint(
        #     scope=self,
        #     id='cfn-efs-access-point',
        #     file_system_id=md_rest_efs_id,
        #     posix_user=efs.CfnAccessPoint.PosixUserProperty(
        #         gid='993',
        #         uid='995'
        #     ),
        #     root_directory=efs.CfnAccessPoint.RootDirectoryProperty(
        #         path=md_rest_efs_folder_name
        #     ),
        #     access_point_tags=[
        #         efs.CfnAccessPoint.AccessPointTagProperty(
        #             key='Name',
        #             value=f'{selected_environment}-{MULTICARRIER_EMAIL_ID}'
        #         ),
        #         efs.CfnAccessPoint.AccessPointTagProperty(
        #             key='Environment',
        #             value=selected_environment
        #         )
        #     ]
        # )

        # cfn_efs_access_point.AccessPointTagProperty(
        #     key='Name',
        #     value=f'{selected_environment}-{MULTICARRIER_EMAIL_ID}-rest'
        # )

        md_rest_access_point_arn = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/efs/access-points/{MULTICARRIER_EMAIL_ID}-multidialogo-rest/arn',
        )

        task_definition = ecs.FargateTaskDefinition(
            scope=self,
            id=f'{service_name}-task-definition',
            cpu=int(service_cpu),
            family=f'{selected_environment}-{service_name}',
            memory_limit_mib=int(service_memory)
        )

        task_definition.apply_removal_policy(
            policy=RemovalPolicy.RETAIN_ON_UPDATE_OR_DELETE
        )

        task_definition.add_to_execution_role_policy(
            statement=iam.PolicyStatement(
                actions=[
                    'elasticfilesystem:ClientMount'
                ],
                resources=[
                    md_rest_access_point_arn
                ]
            )
        )

        mc_email_access_point_arn = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/efs/access-points/multicarrier-email/arn',
        )

        mc_email_access_point_id = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/efs/access-points/multicarrier-email/id',
        )

        mc_email_efs_id = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/efs/file-systems/multicarrier-email/id',
        )

        task_definition.add_to_execution_role_policy(
            statement=iam.PolicyStatement(
                actions=[
                    'elasticfilesystem:ClientMount',
                    'elasticfilesystem:ClientWrite',
                    'elasticfilesystem:ClientRootAccess'
                ],
                resources=[
                    mc_email_access_point_arn
                ]
            )
        )

        md_rest_efs_id = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/efs/file-systems/multidialogo-rest/id',
        )

        md_rest_access_point_id = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/efs/access-points/{MULTICARRIER_EMAIL_ID}-multidialogo-rest/id',
        )

        task_definition.add_volume(
            name=MD_REST_VOLUME_NAME,
            efs_volume_configuration=ecs.EfsVolumeConfiguration(
                file_system_id=md_rest_efs_id,
                transit_encryption='ENABLED',
                authorization_config=ecs.AuthorizationConfig(
                    access_point_id=md_rest_access_point_id,
                    iam='ENABLED'
                )
            )
        )

        task_definition.add_volume(
            name=MC_VOLUME_NAME,
            efs_volume_configuration=ecs.EfsVolumeConfiguration(
                file_system_id=mc_email_efs_id,
                transit_encryption='ENABLED',
                authorization_config=ecs.AuthorizationConfig(
                    access_point_id=mc_email_access_point_id,
                    iam='ENABLED'
                )
            )
        )

        log_group_retainment = RemovalPolicy.RETAIN if selected_environment == 'prod' else RemovalPolicy.DESTROY

        log_group = logs.LogGroup(
            scope=self,
            id=f'{service_name}-log-group',
            log_group_name=f'{selected_environment}/{MULTICARRIER_EMAIL_ID}/{service_name}2',
            removal_policy=log_group_retainment,
            retention=logs.RetentionDays.ONE_MONTH
        )

        repository_name = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/ecr/repositories/{MULTICARRIER_EMAIL_ID}-api/name',
        )

        repository = ecr.Repository.from_repository_name(
            scope=self,
            id='ecr-repository',
            repository_name=repository_name,
            # repository_name=f'{selected_environment}-{service_name}',
        )

        container = task_definition.add_container(
            id=service_name,
            image=ecs.ContainerImage.from_ecr_repository(
                repository=repository,
                tag=image_tag
            ),
            logging=ecs.LogDriver.aws_logs(
                stream_prefix=f'{selected_environment}/{service_name}',
                log_group=log_group
            )
        )

        container.add_environment(
            name='ATTACHMENTS_BASE_PATH',
            value=md_rest_efs_folder_name
        )

        container.add_environment(
            name='EML_STORAGE_PATH',
            value=mc_email_efs_folder_name
        )

        container.add_port_mappings(
            ecs.PortMapping(
                container_port=int(service_container_port),
                host_port=int(service_host_port)
            )
        )

        container.add_mount_points(
            ecs.MountPoint(
                container_path=md_rest_efs_folder_name,
                source_volume=MD_REST_VOLUME_NAME,
                read_only=True
            ),
            ecs.MountPoint(
                container_path=mc_email_efs_folder_name,
                source_volume=MC_VOLUME_NAME,
                read_only=False
            )
        )

        # table_arn = ssm.StringParameter.value_from_lookup(
        #     scope=self,
        #     parameter_name=f'/{selected_environment}/dynamodb/tables/{MULTICARRIER_EMAIL_ID}-outbox/arn',
        # )

        table_name = ssm.StringParameter.value_from_lookup(
            scope=self,
            parameter_name=f'/{selected_environment}/dynamodb/tables/{MULTICARRIER_EMAIL_ID}-outbox/name',
        )

        table = dynamodb.Table.from_table_name(
            scope=self,
            id='table',
            table_name=table_name,
            # table_name=f'{selected_environment}-{MULTICARRIER_EMAIL_ID}-outbox',
        )

        table.grant_read_write_data(
            grantee=task_definition.task_role
        )

        table.grant(
            task_definition.task_role,
            'dynamodb:PartiQLSelect',
            'dynamodb:PartiQLInsert',
            'dynamodb:PartiQLUpdate',
            'dynamodb:PartiQLDelete'
        )

        container.add_environment(
            name='EMAIL_OUTBOX_TABLE',
            value=table.table_name
        )

        # vpc = ec2.Vpc.from_lookup(
        #     scope=self,
        #     id='vpc',
        #     vpc_id=vpc_id
        # )

        # cluster = ecs.Cluster.from_cluster_attributes(
        #     scope=self,
        #     id='cluster',
        #     cluster_name=cluster_name,
        #     vpc=vpc
        # )

        # security_group = ec2.SecurityGroup(
        #     scope=self,
        #     id=f'{service_name}-security-group',
        #     vpc=vpc
        # )
        #
        # md_efs_security_group = ec2.SecurityGroup.from_lookup_by_id(
        #     scope=self,
        #     id='md-efs-security-group',
        #     security_group_id=md_rest_efs_security_group_id
        # )

        # mc_email_efs_security_group_id = ssm.StringParameter.value_from_lookup(
        #         scope=self,
        #         parameter_name=f'/{selected_environment}/ec2/file-systems/{MULTICARRIER_EMAIL_ID}/security-group-id'
        # )

        # mc_email_efs_security_group = ec2.SecurityGroup.from_lookup_by_id(
        #     scope=self,
        #     id='mc-efs-security-group',
        #     security_group_id=mc_email_efs_security_group_id
        # )

        # mc_email_efs_security_group.add_ingress_rule(
        #     peer=ec2.Peer.security_group_id(
        #         security_group_id=security_group.security_group_id,
        #     ),
        #     connection=ec2.Port.NFS
        # )

        # md_efs_security_group.add_ingress_rule(
        #     peer=ec2.Peer.security_group_id(
        #         security_group_id=security_group.security_group_id,
        #     ),
        #     connection=ec2.Port.NFS
        # )

        # listener_arn = ssm.StringParameter.value_from_lookup(
        #         scope=self,
        #         parameter_name=f'/{selected_environment}/ec2/load_balancers/internal-alb/listeners/http/arn'
        # )

        # alb_security_group_id = ssm.StringParameter.value_from_lookup(
        #         scope=self,
        #         parameter_name=f'/{selected_environment}/ec2/load_balancers/internal-alb/security-group-id'
        # )

        # alb_security_group = ec2.SecurityGroup.from_lookup_by_id(
        #     scope=self,
        #     id='internal-alb-security-group',
        #     security_group_id=alb_security_group_id
        # )

        # alb_security_group.add_egress_rule(
        #     peer=ec2.Peer.security_group_id(
        #         security_group_id=security_group.security_group_id,
        #     ),
        #     connection=ec2.Port.HTTP
        # )

        # security_group.add_ingress_rule(
        #     peer=ec2.Peer.security_group_id(
        #         security_group_id=alb_security_group.security_group_id,
        #     ),
        #     connection=ec2.Port.HTTP
        # )

        # service = ecs.FargateService(
        #     scope=self,
        #     id=f'{service_name}-service',
        #     assign_public_ip=False,
        #     cluster=cluster,
        #     desired_count=service_desired_count,
        #     enable_ecs_managed_tags=True,
        #     enable_execute_command=True,
        #     security_groups=[
        #         security_group
        #     ],
        #     service_name=service_name,
        #     task_definition=task_definition,
        #     min_healthy_percent=100,
        #     vpc_subnets=ec2.SubnetSelection(
        #         subnet_type=ec2.SubnetType.PRIVATE_WITH_EGRESS,
        #         one_per_az=True
        #     )
        # )
        #
        # target_group = elbv2.ApplicationTargetGroup(
        #     scope=self,
        #     id=f'{service_name}-target-group',
        #     port=80,
        #     protocol=elbv2.ApplicationProtocol.HTTP,
        #     vpc=vpc,
        #     targets=[
        #         service
        #     ],
        #     target_type=elbv2.TargetType.IP,
        #     health_check=elbv2.HealthCheck(
        #         enabled=True,
        #         path='/health-check',
        #         port=str(service_host_port),
        #     ),
        #     deregistration_delay=Duration.seconds(30)
        # )
        #
        # alb_http_listener = elbv2.ApplicationListener.from_application_listener_attributes(
        #     scope=self,
        #     id='internal-alb-http-listener',
        #     listener_arn=listener_arn,
        #     security_group=alb_security_group,
        #     default_port=80
        # )
        #
        # hosted_zone = route53.PrivateHostedZone(
        #     scope=self,
        #     id='hosted-zone',
        #     vpc=vpc,
        #     zone_name=f'{MULTICARRIER_EMAIL_ID}.{selected_environment}.{md_domain}'
        # )
        #
        # internal_alb_arn = ssm.StringParameter.value_from_lookup(
        #     scope=self,
        #     parameter_name=f'/{selected_environment}/ec2/load_balancers/internal-alb/arn'
        # )
        #
        # alb = elbv2.ApplicationLoadBalancer.from_lookup(
        #     scope=self,
        #     id='internal-alb',
        #     load_balancer_arn=internal_alb_arn
        # )
        #
        # a_record = route53.ARecord(
        #     scope=self,
        #     id='dns-record',
        #     zone=hosted_zone,
        #     record_name='',
        #     target=route53.RecordTarget.from_alias(
        #         route53_targets.LoadBalancerTarget(
        #             alb
        #         )
        #     )
        # )
        #
        # alb_http_listener.add_target_groups(
        #     id=f'{service_name}-target-group-attachment',
        #     target_groups=[
        #         target_group
        #     ],
        #     conditions=[
        #         elbv2.ListenerCondition.path_patterns(
        #             [
        #                 '/emails'
        #             ]
        #         ),
        #         elbv2.ListenerCondition.host_headers(
        #             [
        #                 a_record.domain_name
        #             ]
        #         )
        #     ],
        #     priority=alb_rule_priority
        # )
        #
        # multicarrier_email_daemon_scalable_target = service.auto_scale_task_count(
        #     min_capacity=0,
        #     max_capacity=service_max_count
        # )
        #
        # multicarrier_email_daemon_scalable_target.scale_on_cpu_utilization(
        #     id='cpu-scaling',
        #     target_utilization_percent=70,
        #     scale_in_cooldown=Duration.seconds(60),
        #     scale_out_cooldown=Duration.seconds(60)
        # )
        #
        # multicarrier_email_daemon_scalable_target.scale_on_memory_utilization(
        #     id='memory-scaling',
        #     target_utilization_percent=75,
        #     scale_in_cooldown=Duration.seconds(60),
        #     scale_out_cooldown=Duration.seconds(60)
        # )
        #
        # multicarrier_email_daemon_scalable_target.scale_on_schedule(
        #     id='scale-up-8am',
        #     schedule=applicationautoscaling.Schedule.cron(
        #         minute='0',
        #         hour='8'
        #     ),
        #     min_capacity=service_desired_count,
        #     max_capacity=service_max_count,
        #     time_zone=TimeZone.EUROPE_ROME
        # )
        #
        # multicarrier_email_daemon_scalable_target.scale_on_schedule(
        #     id='scale-down-at-20',
        #     schedule=applicationautoscaling.Schedule.cron(
        #         minute='0',
        #         hour='20'
        #     ),
        #     min_capacity=0,
        #     max_capacity=0,
        #     time_zone=TimeZone.EUROPE_ROME
        # )