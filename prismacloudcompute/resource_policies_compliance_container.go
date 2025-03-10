package prismacloudcompute

import (
	"fmt"
	"time"

	"github.com/paloaltonetworks/prisma-cloud-compute-go/collection"
	"github.com/paloaltonetworks/prisma-cloud-compute-go/pcc"
	"github.com/paloaltonetworks/prisma-cloud-compute-go/policy"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourcePoliciesComplianceContainer() *schema.Resource {
	return &schema.Resource{
		Create: createPolicyComplianceContainer,
		Read:   readPolicyComplianceContainer,
		Update: updatePolicyComplianceContainer,
		Delete: deletePolicyComplianceContainer,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of policy rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"block_message": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Message to display for blocked requests.",
						},
						"collections": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of collections used to scope the rule.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"conditions": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "The set of compliance checks.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_check": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "A compliance check. Omitted compliance checks are ignored.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"block": {
													Type:        schema.TypeBool,
													Optional:    true,
													Description: "Whether or not to block if this check is failed. Setting to 'false' will only alert on failure.",
												},
												"id": {
													Type:        schema.TypeInt,
													Optional:    true,
													Description: "Compliance check ID.",
												},
											},
										},
									},
								},
							},
						},
						"disabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether or not to disable the rule.",
						},
						"effect": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The effect of the rule. Can be set to 'ignore', 'alert', 'block', or 'alert, block'.",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Unique name of the rule.",
						},
						"notes": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Free-form text field.",
						},
						"show_passed_checks": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether or not to report both failed and passed compliance checks.",
						},
						"verbose": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether or not to provide verbose output for blocked requests.",
						},
					},
				},
			},
		},
	}
}

func createPolicyComplianceContainer(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pcc.Client)
	parsedPolicy, err := parseComplianceContainerPolicy(d)
	if err != nil {
		return fmt.Errorf("error creating %s policy: %s", policyTypeComplianceContainer, err)
	}

	if err := policy.UpdateComplianceContainer(*client, *parsedPolicy); err != nil {
		return fmt.Errorf("error creating %s policy: %s", policyTypeComplianceContainer, err)
	}

	d.SetId(policyTypeComplianceContainer)
	return readPolicyComplianceContainer(d, meta)
}

func readPolicyComplianceContainer(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pcc.Client)
	retrievedPolicy, err := policy.GetComplianceContainer(*client)
	if err != nil {
		return fmt.Errorf("error reading %s policy: %s", policyTypeComplianceContainer, err)
	}

	if err := d.Set("rule", flattenPolicyComplianceContainerRules(retrievedPolicy.Rules)); err != nil {
		return fmt.Errorf("error reading %s policy: %s", policyTypeComplianceContainer, err)
	}
	return nil
}

func updatePolicyComplianceContainer(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*pcc.Client)
	parsedPolicy, err := parseComplianceContainerPolicy(d)
	if err != nil {
		return fmt.Errorf("error updating %s policy: %s", policyTypeComplianceContainer, err)
	}

	if err := policy.UpdateComplianceContainer(*client, *parsedPolicy); err != nil {
		return fmt.Errorf("error updating %s policy: %s", policyTypeComplianceContainer, err)
	}

	return readPolicyComplianceContainer(d, meta)
}

func deletePolicyComplianceContainer(d *schema.ResourceData, meta interface{}) error {
	// TODO: reset to default policy
	return nil
}

func parseComplianceContainerPolicy(d *schema.ResourceData) (*policy.CompliancePolicy, error) {
	parsedPolicy := policy.CompliancePolicy{
		Type:  policyTypeComplianceContainer,
		Rules: make([]policy.ComplianceRule, 0),
	}
	if rules, ok := d.GetOk("rule"); ok {
		rulesList := rules.([]interface{})
		parsedRules := make([]policy.ComplianceRule, 0, len(rulesList))
		for _, val := range rulesList {
			rule := val.(map[string]interface{})
			parsedRule := policy.ComplianceRule{}

			parsedRule.BlockMessage = rule["block_message"].(string)

			collectionsList := rule["collections"].([]interface{})
			parsedCollections := make([]collection.Collection, 0, len(collectionsList))
			for _, val := range collectionsList {
				parsedCollection := collection.Collection{
					Name: val.(string),
				}
				parsedCollections = append(parsedCollections, parsedCollection)
			}
			parsedRule.Collections = parsedCollections

			conditionsList := rule["conditions"].([]interface{})
			parsedConditions := policy.ComplianceConditions{}
			for _, val := range conditionsList {
				condition := val.(map[string]interface{})
				complianceChecksList := condition["compliance_check"].([]interface{})
				parsedComplianceChecks := make([]policy.ComplianceCheck, 0, len(complianceChecksList))
				for _, val := range complianceChecksList {
					complianceCheck := val.(map[string]interface{})
					parsedComplianceCheck := policy.ComplianceCheck{
						Block: complianceCheck["block"].(bool),
						Id:    complianceCheck["id"].(int),
					}
					parsedComplianceChecks = append(parsedComplianceChecks, parsedComplianceCheck)
				}
				parsedConditions.Checks = parsedComplianceChecks
			}
			parsedRule.Conditions = parsedConditions

			parsedRule.Disabled = rule["disabled"].(bool)
			parsedRule.Effect = rule["effect"].(string)
			parsedRule.Name = rule["name"].(string)
			parsedRule.Notes = rule["notes"].(string)
			parsedRule.ShowPassedChecks = rule["show_passed_checks"].(bool)
			parsedRule.Verbose = rule["verbose"].(bool)

			parsedRules = append(parsedRules, parsedRule)
		}
		parsedPolicy.Rules = parsedRules
	}
	return &parsedPolicy, nil
}

func flattenPolicyComplianceContainerRules(in []policy.ComplianceRule) []interface{} {
	ans := make([]interface{}, 0, len(in))
	for _, val := range in {
		m := make(map[string]interface{})
		m["block_message"] = val.BlockMessage
		m["collections"] = flattenCollections(val.Collections)
		m["conditions"] = flattenComplianceConditions(val.Conditions)
		m["disabled"] = val.Disabled
		m["effect"] = val.Effect
		m["name"] = val.Name
		m["notes"] = val.Notes
		m["show_passed_checks"] = val.ShowPassedChecks
		m["verbose"] = val.Verbose
		ans = append(ans, m)
	}
	return ans
}
