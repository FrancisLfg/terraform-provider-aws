package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsWafv2RegexPatternSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsWafv2RegexPatternSetCreate,
		Read:   resourceAwsWafv2RegexPatternSetRead,
		Update: resourceAwsWafv2RegexPatternSetUpdate,
		Delete: resourceAwsWafv2RegexPatternSetDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"regular_expression_list": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"scope": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					wafv2.ScopeCloudfront,
					wafv2.ScopeRegional,
				}, false),
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsWafv2RegexPatternSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.CreateRegexPatternSetOutput

	log.Printf("[INFO] Creating WAFV2 Regex Pattern Set: %s", d.Get("name").(string))

	params := &wafv2.CreateRegexPatternSetInput{
		Name:                  aws.String(d.Get("name").(string)),
		RegularExpressionList: []*wafv2.Regex{},
		Scope:                 aws.String(d.Get("scope").(string)),
	}

	if v, ok := d.GetOk("regular_expression_list"); ok && v.(*schema.Set).Len() > 0 {
		var regex []*wafv2.Regex
		for _, r := range expandStringSet(d.Get("regular_expression_list").(*schema.Set)) {
			regex = append(regex, &wafv2.Regex{RegexString: r})
		}
		params.RegularExpressionList = regex
	}
	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		params.Tags = keyvaluetags.New(v).IgnoreAws().Wafv2Tags()
	}

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.CreateRegexPatternSet(params)
		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAFV2 couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationException, "An error occurred during the tagging operation") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationInternalErrorException, "AWS WAFV2 couldn’t perform your tagging operation because of an internal error") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAFV2 couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if isResourceTimeoutError(err) {
		_, err = conn.CreateRegexPatternSet(params)
	}

	if err != nil {
		return fmt.Errorf("failed creating WAFV2 Regex Pattern Set: %s", err)
	}
	d.SetId(*resp.Summary.Id)

	return resourceAwsWafv2RegexPatternSetRead(d, meta)
}

func resourceAwsWafv2RegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Reading WAFv2 Regex Pattern Set: %s", d.Get("name").(string))
	params := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}

	resp, err := conn.GetRegexPatternSet(params)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == wafv2.ErrCodeWAFNonexistentItemException {
			log.Printf("[WARN] WAFV2 RegexPatternSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", resp.RegexPatternSet.Name)
	d.Set("description", resp.RegexPatternSet.Description)
	d.Set("regular_expression_list", resp.RegexPatternSet.RegularExpressionList)
	d.Set("arn", resp.RegexPatternSet.ARN)

	var regexList []*string
	for _, v := range resp.RegexPatternSet.RegularExpressionList {
		regexList = append(regexList, v.RegexString)
	}

	if err := d.Set("regular_expression_list", schema.NewSet(schema.HashString, flattenStringList(regexList))); err != nil {
		return fmt.Errorf("error setting regular_expression_list: %s", err)
	}

	tags, err := keyvaluetags.Wafv2ListTags(conn, *resp.RegexPatternSet.ARN)
	if err != nil {
		return fmt.Errorf("error listing tags for WAFV2 Regex Pattern Set (%s): %s", *resp.RegexPatternSet.ARN, err)
	}

	if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsWafv2RegexPatternSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn
	var resp *wafv2.GetRegexPatternSetOutput
	params := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}
	log.Printf("[INFO] Updating WAFV2 Regex Pattern Set %s", d.Id())

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.GetRegexPatternSet(params)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("error getting lock token: %s", err))
		}

		u := &wafv2.UpdateRegexPatternSetInput{
			Id:                    aws.String(d.Id()),
			Name:                  aws.String(d.Get("name").(string)),
			Scope:                 aws.String(d.Get("scope").(string)),
			RegularExpressionList: []*wafv2.Regex{},
			Description:           aws.String(d.Get("description").(string)),
			LockToken:             resp.LockToken,
		}

		if v, ok := d.GetOk("regular_expression_list"); ok && v.(*schema.Set).Len() > 0 {
			var regex []*wafv2.Regex
			for _, r := range expandStringSet(d.Get("regular_expression_list").(*schema.Set)) {
				regex = append(regex, &wafv2.Regex{RegexString: r})
			}
			u.RegularExpressionList = regex
		}

		if d.HasChange("description") {
			u.Description = aws.String(d.Get("description").(string))
		}

		_, err = conn.UpdateRegexPatternSet(u)

		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAFV2 couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAFV2 couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRegexPatternSet(&wafv2.DeleteRegexPatternSetInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})
	}

	if err != nil {
		return fmt.Errorf("error updating WAFV2 Regex Pattern Set: %s", err)
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Wafv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceAwsWafv2RegexPatternSetRead(d, meta)
}

func resourceAwsWafv2RegexPatternSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	var resp *wafv2.GetRegexPatternSetOutput
	params := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(d.Id()),
		Name:  aws.String(d.Get("name").(string)),
		Scope: aws.String(d.Get("scope").(string)),
	}
	log.Printf("[INFO] Deleting WAFV2 RegexPatternSet %s", d.Id())

	err := resource.Retry(15*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = conn.GetRegexPatternSet(params)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting lock token: %s", err))
		}

		_, err = conn.DeleteRegexPatternSet(&wafv2.DeleteRegexPatternSetInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})

		if err != nil {
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAFV2 couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAFV2 couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.DeleteRegexPatternSet(&wafv2.DeleteRegexPatternSetInput{
			Id:        aws.String(d.Id()),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
			LockToken: resp.LockToken,
		})
	}

	if err != nil {
		return fmt.Errorf("error deleting WAFV2 Regex Pattern Set: %s", err)
	}

	return nil
}
