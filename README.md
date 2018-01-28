# AWS tools

Some specialised tools to avoid pulling boto3

* `cloudwatch-put-metric-data`: Basic sending a metric value to cloudwatch
* `ec2-ip-from-name`: Given an EC2 name, list up to `-max-results` IPs associated with instances with that name (default is 1).
* `ecs-deploy`: Update the container images of a task and update services to use it
* `ecs-run-task`: Run a task definition
* `elb-resolve-elb-external-url`: ELB classic only (no ALB). Given a name returns the zone53 record associated with the ELB, including scheme (https returned if both available) and port.
* `elb-resolve-alb-external-url`: Both ELB classic and ALB. Given a name, returns route53 record associated with the ELB. Does not include scheme or port as it doesn't check listeners.
* `s3-download`: Download a single file from s3
