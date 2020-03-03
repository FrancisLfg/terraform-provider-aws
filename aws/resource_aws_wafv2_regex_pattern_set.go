package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
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
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 256),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"regular_expression_list": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeMap},
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

	log.Printf("[INFO] Creating WAF V2 Regex Pattern Set: %s", d.Get("name").(string))

	params := &wafv2.CreateRegexPatternSetInput{
		Description:           aws.String(d.Get("description").(string)),
		Name:                  aws.String(d.Get("name").(string)),
		RegularExpressionList: []*wafv2.Regex{},
		Scope:                 aws.String(d.Get("scope").(string)),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		//params.RegularExpressionList = d.Get("regular_expression_list").(*schema.Set)
		params.RegularExpressionList = nil
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
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationException, "An error occurred during the tagging operation") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFTagOperationInternalErrorException, "AWS WAF couldn’t perform your tagging operation because of an internal error") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
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
		return fmt.Errorf("Failed creating WAF Regex Pattern Set: %s", err)
	}
	d.SetId(*resp.Summary.Id)

	return resourceAwsWafv2RegexPatternSetUpdate(d, meta)
}

func resourceAwsWafv2RegexPatternSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).wafv2conn

	log.Printf("[INFO] Reading WAF Regional Regex Pattern Set: %s", d.Get("name").(string))
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

	//if err := d.Set("addresses", schema.NewSet(schema.HashString, flattenStringList(resp.IPSet.Addresses))); err != nil {
	//	return fmt.Errorf("Error setting addresses: %s", err)
	//}
	//
	//tags, err := keyvaluetags.Wafv2ListTags(conn, *resp.IPSet.ARN)
	//if err != nil {
	//	return fmt.Errorf("error listing tags for WAFV2 IpSet (%s): %s", *resp.IPSet.ARN, err)
	//}
	//
	//if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
	//	return fmt.Errorf("error setting tags: %s", err)
	//}
	//

	return nil
}

func resourceAwsWafv2RegexPatternSetUpdate(d *schema.ResourceData, meta interface{}) error {
	//conn := meta.(*AWSClient).wafv2conn
	//
	//log.Printf("[INFO] Updating WAF Regex Pattern Set: %s", d.Get("name").(string))

	//if d.HasChange("regex_pattern_strings") {
	//	o, n := d.GetChange("regex_pattern_strings")
	//	oldPatterns, newPatterns := o.(*schema.Set).List(), n.(*schema.Set).List()
	//	err := updateWafv2RegexPatternSetPatternStringsWR(d.Id(), oldPatterns, newPatterns, conn, region)
	//	if isAWSErr(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
	//		log.Printf("[WARN] WAF Regional Rate Based Rule (%s) not found, removing from state", d.Id())
	//		d.SetId("")
	//		return nil
	//	}
	//	if err != nil {
	//		return fmt.Errorf("Failed updating WAF Regional Regex Pattern Set(%s): %s", d.Id(), err)
	//	}
	//}

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
			if isAWSErr(err, wafv2.ErrCodeWAFInternalErrorException, "AWS WAF couldn’t perform the operation because of a system problem") {
				return resource.RetryableError(err)
			}
			if isAWSErr(err, wafv2.ErrCodeWAFOptimisticLockException, "AWS WAF couldn’t save your changes because you tried to update or delete a resource that has changed since you last retrieved it") {
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
		return fmt.Errorf("Error deleting WAFV2 IPSet: %s", err)
	}

	return nil
}

//func updateWafv2RegexPatternSetPatternStringsWR(id string, oldPatterns, newPatterns []interface{}, conn *wafregional.WAFRegional) error {
//	wr := newWafv2Retryer(conn)
//	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
//		req := &waf.UpdateRegexPatternSetInput{
//			ChangeToken:       token,
//			RegexPatternSetId: aws.String(id),
//			Updates:           diffWafRegexPatternSetPatternStrings(oldPatterns, newPatterns),
//		}
//
//		return conn.UpdateRegexPatternSet(req)
//	})
//
//	return err
//}
