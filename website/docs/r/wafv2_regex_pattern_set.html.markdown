---
subcategory: "WAFv2"
layout: "aws"
page_title: "AWS: aws_wafv2_regex_pattern_set"
description: |-
  Provides an AWS WAFv2 Regex Pattern Set resource.
---

# Resource: aws_wafv2_regex_pattern_set

Provides a WAFv2 Regex Pattern Set Resource

## Example Usage

```hcl
resource "aws_wafv2_regex_pattern_set" "test" {
  name                    = "example"
  description             = "Example regex pattern set"
  scope                   = "REGIONAL"
  regular_expression_list = ["^foobar$","^example$"]
  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) A friendly name of the set.
* `description` - (Optional) A friendly description of the set.
* `scope` - (Required) Specifies whether this is for an AWS CloudFront distribution or for a regional application. Valid values are `CLOUDFRONT` or `REGIONAL`. To work with CloudFront, you must also specify the Region US East (N. Virginia).
* `regular_expression_list` - (Required) Array of regular expression strings.
* `tags` - (Optional) An array of key:value pairs to associate with the resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - A unique identifier for the set.
* `arn` - The Amazon Resource Name (ARN) that identifies the cluster.

## Import

WAFv2 Regex Pattern Sets can be imported using their ID, Name and Scope e.g.

```
$ terraform import aws_wafv2_regex_pattern_set.example a1b2c3d4-d5f6-7777-8888-9999aaaabbbbcccc/example/REGIONAL