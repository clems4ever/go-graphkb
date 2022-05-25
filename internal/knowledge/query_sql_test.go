package knowledge

import (
	"strings"
	"testing"

	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type QueryCase struct {
	Description string
	Cypher      string
	SQL         string
	Error       string
	Selected    bool
}

func TestQueryTranslation(t *testing.T) {
	cases := []QueryCase{
		{
			Cypher: "MATCH (n:ip) RETURN n",
			SQL:    `SELECT a0.id, a0.value, a0.type FROM (assets ip) JOIN assets a0 ON a0.type = 'ip' AND a0.id = ip.id`,
		},
		{
			Cypher: "MATCH (n:ip), (n:name) RETURN n",
			Error:  "Variable 'n' already defined with a different type",
		},
		{
			Cypher: "MATCH (n:ip) RETURN n, n",
			SQL:    `SELECT a0.id, a0.value, a0.type, a0.id, a0.value, a0.type FROM (assets ip) JOIN assets a0 ON a0.type = 'ip' AND a0.id = ip.id`,
		},
		{
			Cypher: "MATCH (n) WHERE n.value = 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM (assets a0) WHERE a0.value = 'prod'",
		},
		{
			Cypher: "MATCH (n) WHERE NOT n.value = 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM (assets a0) WHERE NOT a0.value = 'prod'",
		},
		{
			Cypher: "MATCH (n) WHERE NOT n.value = 'prod' AND n.value = 'preprod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM (assets a0) WHERE NOT a0.value = 'prod' AND a0.value = 'preprod'",
		},
		{
			Cypher: "MATCH (n) WHERE n.value STARTS WITH 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM (assets a0) WHERE a0.value LIKE 'prod%'",
		},
		{
			Cypher: "MATCH (n) WHERE n.value ENDS WITH 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM (assets a0) WHERE a0.value LIKE '%prod'",
		},
		{
			Cypher: "MATCH (n) WHERE n.value CONTAINS 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM (assets a0) WHERE a0.value LIKE '%prod%'",
		},
		{
			Cypher: "MATCH (:variable)-[:has]->(n:name) RETURN n",
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id`,
		},
		{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n",
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id`,
		},
		{
			Cypher: "MATCH (v:variable)--(n:name) RETURN n",
			SQL: `
			(SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id) 
			UNION ALL 
			(SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id)`,
		},
		{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN n",
			SQL: `
			(SELECT a1.id, a1.value, a1.type 
			FROM (assets variable)
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id) 
			UNION ALL 
			(SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id)`,
		},
		{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN n LIMIT 10",
			SQL: `
			(SELECT a1.id, a1.value, a1.type 
			FROM (assets variable)
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id) 
			UNION ALL 
			(SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id)
			LIMIT 10`,
		},
		{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN v.value, COUNT(n.value)",
			SQL: `
			SELECT a0_value, SUM(a1_value_COUNT) 
			FROM (
				(SELECT a0.value AS a0_value, COUNT(a1.value) AS a1_value_COUNT 
				FROM (assets variable) 
				JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
				JOIN relations r0 ON r0.from_id = a0.id 
				JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id 
				GROUP BY a0_value) 
				UNION ALL 
				(SELECT a0.value AS a0_value, COUNT(a1.value) AS a1_value_COUNT 
				FROM (assets variable) 
				JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
				JOIN relations r0 ON r0.to_id = a0.id 
				JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id 
				GROUP BY a0_value)
			) AS x GROUP BY x.a0_value`,
		},
		{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN DISTINCT n.value LIMIT 10",
			SQL: `
			(SELECT a1.value 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id) 
			UNION 
			(SELECT a1.value 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id) LIMIT 10`,
		},
		{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN v.value, COUNT(DISTINCT n.value)",
			SQL: `
			SELECT a0_value, SUM(a1_value_COUNT) 
			FROM (
				(SELECT a0.value AS a0_value, COUNT(DISTINCT a1.value) AS a1_value_COUNT 
				FROM (assets variable) 
				JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
				JOIN relations r0 ON r0.from_id = a0.id 
				JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id 
				GROUP BY a0_value) 
				UNION ALL 
				(SELECT a0.value AS a0_value, COUNT(DISTINCT a1.value) AS a1_value_COUNT 
				FROM (assets variable) 
				JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
				JOIN relations r0 ON r0.to_id = a0.id 
				JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id 
				GROUP BY a0_value)
			) AS x GROUP BY x.a0_value`,
		},
		{
			Cypher: "MATCH (v)-[r]-(n) RETURN n LIMIT 10",
			SQL: `
			SELECT a1.id, a1.value, a1.type FROM (assets a0) JOIN relations r0 ON r0.to_id = a0.id JOIN assets a1 ON r0.from_id = a1.id LIMIT 10`,
		},
		{
			Cypher: "MATCH (v:variable)<-[r]-(n:name), (v)-[r1]->(n) RETURN n",
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.to_id = a0.id 
			JOIN relations r1 ON r1.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id AND r1.to_id = a1.id`,
		},
		{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n.value",
			SQL: `
			SELECT a1.value 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id`,
		},
		{
			Cypher: "MATCH (v:variable)<-[r:has]-(n:name) RETURN v, r, n",
			SQL: `
			SELECT a0.id, a0.value, a0.type, r0.id, r0.from_id, r0.to_id, r0.type, a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id`,
		},
		{
			Cypher: "MATCH (v:variable)<-[r]-(n) RETURN v, r, n",
			SQL: `
			SELECT a0.id, a0.value, a0.type, r0.id, r0.from_id, r0.to_id, r0.type, a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.to_id = a0.id 
			JOIN assets a1 ON r0.from_id = a1.id`,
		},
		{
			Cypher: "MATCH (v:variable)<-[:has]-(:name)-[:is_in]->(:program) RETURN v",
			SQL: `
			SELECT a0.id, a0.value, a0.type
			FROM (assets variable)
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id
			JOIN relations r1 ON r1.type = 'is_in' AND r1.from_id = a1.id
			JOIN assets a2 ON a2.type = 'program' AND r1.to_id = a2.id`,
		},
		{
			Cypher: `MATCH (p:port)<-[:bind]-(c:consul_service)-[:is_in]->(d:datacenter) WHERE d.value = 'pa4'
					MATCH (c)-[:is_in]->(e:environment) WHERE e.value = 'preprod'
					RETURN c`,
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets port) 
			JOIN assets a0 ON a0.type = 'port' AND a0.id = port.id 
			JOIN relations r0 ON r0.type = 'bind' AND r0.to_id = a0.id  
			JOIN assets a1 ON a1.type = 'consul_service' AND r0.from_id = a1.id 
			JOIN relations r1 ON r1.type = 'is_in' AND r1.from_id = a1.id 
			JOIN relations r2 ON r2.type = 'is_in' AND r2.from_id = a1.id 
			JOIN assets a2 ON a2.type = 'datacenter' AND r1.to_id = a2.id 
			JOIN assets a3 ON a3.type = 'environment' AND r2.to_id = a3.id 
			WHERE a2.value = 'pa4' AND a3.value = 'preprod'`,
		},
		{
			Cypher: `MATCH (p:port)<-[:bind]-(c:consul_service)-[:is_in]->(d:datacenter) WHERE d.value = 'pa4'
					MATCH (c)-[:is_in]->(e:environment) WHERE e.value <> 'preprod'
					RETURN c`,
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets port) 
			JOIN assets a0 ON a0.type = 'port' AND a0.id = port.id 
			JOIN relations r0 ON r0.type = 'bind' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'consul_service' AND r0.from_id = a1.id 
			JOIN relations r1 ON r1.type = 'is_in' AND r1.from_id = a1.id 
			JOIN relations r2 ON r2.type = 'is_in' AND r2.from_id = a1.id 
			JOIN assets a2 ON a2.type = 'datacenter' AND r1.to_id = a2.id 
			JOIN assets a3 ON a3.type = 'environment' AND r2.to_id = a3.id 
			WHERE a2.value = 'pa4' AND a3.value <> 'preprod'`,
		},
		{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n LIMIT 10",
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id 
			LIMIT 10`,
		},
		{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n SKIP 20 LIMIT 10",
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id 
			LIMIT 10 
			OFFSET 20`,
		},
		{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN DISTINCT n",
			SQL: `
			SELECT DISTINCT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.from_id = a1.id`,
		},
		{
			Cypher: `MATCH (r:rack)<-[:is_in]-(d:device)-[:is_in]->(e:environment)
					WHERE r.value = '01.04'
					RETURN e.value, COUNT(d.value)`,
			SQL: `
			SELECT a2.value, COUNT(a1.value) 
			FROM (assets rack) 
			JOIN assets a0 ON a0.type = 'rack' AND a0.id = rack.id 
			JOIN relations r0 ON r0.type = 'is_in' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'device' AND r0.from_id = a1.id 
			JOIN relations r1 ON r1.type = 'is_in' AND r1.from_id = a1.id 
			JOIN assets a2 ON a2.type = 'environment' AND r1.to_id = a2.id WHERE a0.value = '01.04' 
			GROUP BY a2.value`,
		},
		{
			Cypher: `MATCH (r:rack)<-[:is_in]-(d:device) RETURN COUNT(d)`,
			SQL: `
			SELECT COUNT(a1.id) 
			FROM (assets rack) 
			JOIN assets a0 ON a0.type = 'rack' AND a0.id = rack.id 
			JOIN relations r0 ON r0.type = 'is_in' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'device' AND r0.from_id = a1.id`,
		},
		{
			Cypher: `MATCH (r:rack)<-[:is_in]-(d:device) RETURN COUNT(d.value)`,
			SQL: `
			SELECT COUNT(a1.value) 
			FROM (assets rack) 
			JOIN assets a0 ON a0.type = 'rack' AND a0.id = rack.id 
			JOIN relations r0 ON r0.type = 'is_in' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'device' AND r0.from_id = a1.id`,
		},
		{
			Cypher: "MATCH (v:variable)-[:has]->(n:name) WHERE v.value = '0x16' AND (n.value = 'myvar' OR n.value = 'myvar2') RETURN n",
			SQL: `
			SELECT a1.id, a1.value, a1.type 
			FROM (assets variable) 
			JOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.from_id = a0.id 
			JOIN assets a1 ON a1.type = 'name' AND r0.to_id = a1.id 
			WHERE (a0.value = '0x16' AND a1.value = 'myvar' OR a1.value = 'myvar2')`,
		},
		{
			Cypher: `
			MATCH (ip:ip)<-[:observed]-(:device)
			WHERE (ip)<-[:has]-(:mesos_task)
			RETURN ip`,
			SQL: `
			SELECT a0.id, a0.value, a0.type 
			FROM (assets ip) JOIN assets a0 ON a0.type = 'ip' AND a0.id = ip.id 
			JOIN relations r0 ON r0.type = 'observed' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'device' AND r0.from_id = a1.id 
			WHERE EXISTS (
				SELECT 1 FROM (assets ip) 
				JOIN assets aw0 ON aw0.type = 'ip' AND aw0.id = a0.id 
				JOIN relations rw0 ON rw0.type = 'has' AND rw0.to_id = aw0.id 
				JOIN assets aw2 ON aw2.type = 'mesos_task' AND rw0.from_id = aw2.id
			)`,
		},
		{
			Cypher: `
			MATCH (ip:ip)<-[:observed]-(:device)
			WHERE NOT (ip)<-[:has]-(:mesos_task)
			RETURN ip`,
			SQL: `
			SELECT a0.id, a0.value, a0.type 
			FROM (assets ip) 
			JOIN assets a0 ON a0.type = 'ip' AND a0.id = ip.id 
			JOIN relations r0 ON r0.type = 'observed' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'device' AND r0.from_id = a1.id 
			WHERE NOT EXISTS (
				SELECT 1 FROM (assets ip) JOIN assets aw0 ON aw0.type = 'ip' AND aw0.id = a0.id 
				JOIN relations rw0 ON rw0.type = 'has' AND rw0.to_id = aw0.id 
				JOIN assets aw2 ON aw2.type = 'mesos_task' AND rw0.from_id = aw2.id)`,
		},
		{
			Description: "Combine a pattern and a simple comparison expression in the WHERE clause",
			Cypher: `
			MATCH (ip:ip)<-[:observed]-(:device)
			WHERE (ip)<-[:has]-(:mesos_task) AND ip.value = '10.244.117.16'
			RETURN ip`,
			SQL: `
			SELECT a0.id, a0.value, a0.type 
			FROM (assets ip) 
			JOIN assets a0 ON a0.type = 'ip' AND a0.id = ip.id 
			JOIN relations r0 ON r0.type = 'observed' AND r0.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'device' AND r0.from_id = a1.id 
			WHERE EXISTS (
				SELECT 1 FROM (assets ip) 
				JOIN assets aw0 ON aw0.type = 'ip' AND aw0.id = a0.id 
				JOIN relations rw0 ON rw0.type = 'has' AND rw0.to_id = aw0.id 
				JOIN assets aw2 ON aw2.type = 'mesos_task' AND rw0.from_id = aw2.id) AND a0.value = '10.244.117.16'`,
		},
		{
			Cypher: `
			MATCH (s:subnet)-[:has]->(r:role) WHERE r.value="LBVIP"
			MATCH (s)-[:is]->(sscope:scope) WHERE sscope.value="public"
			MATCH (f:fqdn)-[:points_to]->(i:ip)-[:is_in]->(s)
			RETURN DISTINCT f
			`,
			SQL: `
			SELECT DISTINCT a3.id, a3.value, a3.type 
			FROM (assets subnet, assets fqdn) 
			JOIN assets a0 ON a0.type = 'subnet' AND a0.id = subnet.id 
			JOIN relations r0 ON r0.type = 'has' AND r0.from_id = a0.id 
			JOIN relations r1 ON r1.type = 'is' AND r1.from_id = a0.id 
			JOIN relations r2 ON r2.type = 'is_in' AND r2.to_id = a0.id 
			JOIN assets a1 ON a1.type = 'role' AND r0.to_id = a1.id 
			JOIN assets a2 ON a2.type = 'scope' AND r1.to_id = a2.id 
			JOIN assets a3 ON a3.type = 'fqdn' AND a3.id = fqdn.id 
			JOIN relations r3 ON r3.type = 'points_to' AND r3.from_id = a3.id 
			JOIN assets a4 ON a4.type = 'ip' AND r3.to_id = a4.id AND r2.from_id = a4.id 
			WHERE a1.value = 'LBVIP' AND a2.value = 'public'
			`,
		},
		{
			Cypher: `
			MATCH (ip:ip)
			MATCH (ip)<-[:has]-(m:mesos_task)
			MATCH (m)-[:has]->(port:port)
			MATCH (m)<-[:has]-(a:marathon_app_version)
			MATCH (a)-[:runs_as]->(s:service_account)
			MATCH (s)-[:has_owner]->(ldap_group:ldap_group)
			RETURN ip, port, ldap_group
			`,
			SQL: `
			SELECT a0.id, a0.value, a0.type, a2.id, a2.value, a2.type, a5.id, a5.value, a5.type
			FROM (assets ip)
			JOIN assets a0 ON a0.type = 'ip' AND a0.id = ip.id
			JOIN relations r0 ON r0.type = 'has' AND r0.to_id = a0.id
			JOIN assets a1 ON a1.type = 'mesos_task' AND r0.from_id = a1.id
			JOIN relations r1 ON r1.type = 'has' AND r1.from_id = a1.id
			JOIN relations r2 ON r2.type = 'has' AND r2.to_id = a1.id
			JOIN assets a2 ON a2.type = 'port' AND r1.to_id = a2.id
			JOIN assets a3 ON a3.type = 'marathon_app_version' AND r2.from_id = a3.id
			JOIN relations r3 ON r3.type = 'runs_as' AND r3.from_id = a3.id
			JOIN assets a4 ON a4.type = 'service_account' AND r3.to_id = a4.id
			JOIN relations r4 ON r4.type = 'has_owner' AND r4.from_id = a4.id
			JOIN assets a5 ON a5.type = 'ldap_group' AND r4.to_id = a5.id 
			`,
		},
	}

	selectionEnabled := false
	for _, c := range cases {
		selectionEnabled = selectionEnabled || c.Selected
	}

	trimFn := func(s string) string {
		c := strings.Join(strings.Fields(s), " ")
		c = strings.ReplaceAll(c, "( ", "(")
		c = strings.ReplaceAll(c, " )", ")")
		return c
	}

	for _, c := range cases {
		if selectionEnabled && !c.Selected {
			continue
		}
		t.Run(c.Cypher, func(t *testing.T) {
			translator := NewSQLQueryTranslator()
			q, err := query.TransformCypher(c.Cypher)
			require.NoError(t, err)

			sql, err := translator.Translate(q)
			if c.Error != "" {
				require.Error(t, err, "Error on test case %s", c.Cypher)
				if err != nil {
					assert.Equal(t, c.Error, err.Error(), "Error on test case %s", c.Cypher)
				}
			} else {
				require.NoError(t, err, "Error on test case %s", c.Cypher)
				expSQL := trimFn(c.SQL)
				actualSQL := trimFn(sql.Query)
				assert.Equal(t, expSQL, actualSQL, "Error on test case %s", c.Cypher)
			}
		})
	}
}

func TestUnwindOrExpressions(t *testing.T) {
	And := func(e ...AndOrExpression) AndOrExpression {
		return AndOrExpression{
			And:      true,
			Children: e,
		}
	}
	Or := func(e ...AndOrExpression) AndOrExpression {
		return AndOrExpression{
			And:      false,
			Children: e,
		}
	}
	Expr := func(e string) AndOrExpression {
		return AndOrExpression{
			Expression: e,
		}
	}

	t.Run("single", func(t *testing.T) {
		exprA := Expr("a")

		unwoundExpr, err := UnwindOrExpressions(exprA)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 1)

		expected := And(exprA)
		assert.Equal(t, expected, unwoundExpr[0])
	})

	t.Run("and", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		expr := And(exprA, exprB)

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 1)

		expected := And(And(exprA), And(exprB))
		assert.Equal(t, expected, unwoundExpr[0])
	})

	t.Run("and_and_and", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		expr := And(exprA, And(exprB, And(exprC, exprD)))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 1)

		expected := And(And(exprA), And(And(exprB), And(And(exprC), And(exprD))))
		assert.Equal(t, expected, unwoundExpr[0])
	})

	t.Run("or", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		expr := Or(exprA, exprB)

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 2)

		assert.Equal(t, And(exprA), unwoundExpr[0])
		assert.Equal(t, And(exprB), unwoundExpr[1])
	})

	t.Run("and_or", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		expr := And(Or(exprA, exprB), Or(exprC, exprD))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 4)

		assert.Equal(t, And(And(exprA), And(exprC)), unwoundExpr[0])
		assert.Equal(t, And(And(exprA), And(exprD)), unwoundExpr[1])
		assert.Equal(t, And(And(exprB), And(exprC)), unwoundExpr[2])
		assert.Equal(t, And(And(exprB), And(exprD)), unwoundExpr[3])
	})

	t.Run("or_and", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		expr := Or(And(exprA, exprB), And(exprC, exprD))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 2)

		assert.Equal(t, And(And(exprA), And(exprB)), unwoundExpr[0])
		assert.Equal(t, And(And(exprC), And(exprD)), unwoundExpr[1])
	})

	t.Run("complex", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		exprE := Expr("e")
		expr := Or(Or(exprA, exprB), And(exprC, Or(exprD, exprE)))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 4)

		assert.Equal(t, And(exprA), unwoundExpr[0])
		assert.Equal(t, And(exprB), unwoundExpr[1])
		assert.Equal(t, And(And(exprC), And(exprD)), unwoundExpr[2])
		assert.Equal(t, And(And(exprC), And(exprE)), unwoundExpr[3])
	})
}
