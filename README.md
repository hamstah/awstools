# AWS tools

Some specialised tools to avoid pulling boto3

* `elb`: ELB classic only (no ALB). Given a name returns the zone53 record associated with the ELB, including scheme (https returned if both available) and port.
* `elb-name`: Both ELB classic and ALB. Given a name, returns route53 record associated with the ELB. Does not include scheme or port as it doesn't check listeners.
* `s3-download`: Download a single file from s3
* `ec2-ip-from-name`: Given an EC2 name, list up to `-max-results` IPs associated with instances with that name (default is 1).
* `ecs`: Run a task definition
