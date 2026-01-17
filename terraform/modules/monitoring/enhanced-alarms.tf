# Application Performance Alarms

resource "aws_cloudwatch_metric_alarm" "api_p99_latency" {
  alarm_name          = "${var.project_name}-${var.environment}-api-p99-latency"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  threshold           = 1000  # 1 second

  metric_query {
    id          = "p99"
    return_data = true

    metric {
      metric_name = "TargetResponseTime"
      namespace   = "AWS/ApplicationELB"
      period      = 60
      stat        = "p99"

      dimensions = {
        LoadBalancer = var.alb_arn_suffix
      }
    }
  }

  alarm_description = "API P99 latency is above 1 second"
  alarm_actions     = [aws_sns_topic.critical_alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "error_rate_spike" {
  alarm_name          = "${var.project_name}-${var.environment}-error-rate-spike"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  threshold           = 5  # 5% error rate

  metric_query {
    id          = "error_rate"
    expression  = "errors / requests * 100"
    label       = "Error Rate"
    return_data = true
  }

  metric_query {
    id = "errors"

    metric {
      metric_name = "HTTPCode_Target_5XX_Count"
      namespace   = "AWS/ApplicationELB"
      period      = 300
      stat        = "Sum"

      dimensions = {
        LoadBalancer = var.alb_arn_suffix
      }
    }
  }

  metric_query {
    id = "requests"

    metric {
      metric_name = "RequestCount"
      namespace   = "AWS/ApplicationELB"
      period      = 300
      stat        = "Sum"

      dimensions = {
        LoadBalancer = var.alb_arn_suffix
      }
    }
  }

  alarm_description = "Error rate is above 5%"
  alarm_actions     = [aws_sns_topic.critical_alerts.arn]
}

# Database Alarms

resource "aws_cloudwatch_metric_alarm" "rds_cpu" {
  alarm_name          = "${var.project_name}-${var.environment}-rds-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    DBInstanceIdentifier = var.rds_identifier
  }

  alarm_description = "RDS CPU utilization is above 80%"
  alarm_actions     = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "rds_connections" {
  alarm_name          = "${var.project_name}-${var.environment}-rds-connections-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    DBInstanceIdentifier = var.rds_identifier
  }

  alarm_description = "RDS connection count is high"
  alarm_actions     = [aws_sns_topic.alerts.arn]
}

# Cost Alarm

resource "aws_cloudwatch_metric_alarm" "estimated_charges" {
  alarm_name          = "${var.project_name}-${var.environment}-cost-alarm"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "EstimatedCharges"
  namespace           = "AWS/Billing"
  period              = 21600
  statistic           = "Maximum"
  threshold           = var.environment == "prod" ? 500 : 150

  dimensions = {
    Currency = "USD"
  }

  alarm_description = "AWS charges exceed threshold"
  alarm_actions     = [aws_sns_topic.critical_alerts.arn]
}

# SNS Topics

resource "aws_sns_topic" "critical_alerts" {
  name_prefix = "${var.project_name}-${var.environment}-critical-"
}

resource "aws_sns_topic_subscription" "critical_email" {
  topic_arn = aws_sns_topic.critical_alerts.arn
  protocol  = "email"
  endpoint  = var.sns_email
}

# Optional: PagerDuty Integration
resource "aws_sns_topic_subscription" "pagerduty" {
  count     = var.pagerduty_endpoint != "" ? 1 : 0
  topic_arn = aws_sns_topic.critical_alerts.arn
  protocol  = "https"
  endpoint  = var.pagerduty_endpoint
}