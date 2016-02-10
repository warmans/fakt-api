package entity
import (
	"strings"
	"time"
)

type Sql struct {
	columns []string
	from    string
	ljoin   []string
	where   []string
	values  []interface{}
	limit   string
	order   []string
}

func (s *Sql) Select(cols... string){
	s.columns = append(s.columns, cols...)
}

func (s *Sql) From(table string){
	s.from = table
}

func (s *Sql) LeftJoin(table... string){
	s.ljoin = append(s.ljoin, table...)
}

func (s *Sql) WhereInt(field, comparison string, value int) {
	s.where = append(s.where, field+" "+comparison+" ?")
	s.values = append(s.values, value)
}

func (s *Sql) WhereStringIn(field string, values... string) {
	if len(values) == 0 {
		return
	}
	s.where = append(s.where, field+" IN (" + (strings.TrimRight(strings.Repeat("?,", len(values)), ",")) + ")")
	s.values = append(s.values, stringsToInterfaces(values)...)
}

func (s *Sql) WhereIntIn(field string, values... int) {
	if len(values) == 0 {
		return
	}
	s.where = append(s.where, field+" IN (" + (strings.TrimRight(strings.Repeat("?,", len(values)), ",")) + ")")
	s.values = append(s.values, intsToInterfaces(values)...)
}

func (s *Sql) WhereTime(field string, comparison string, value time.Time){
	if value.IsZero() {
		return
	}
	s.where = append(s.where, field+" "+comparison+" ?")
	s.values = append(s.values, value.Format(DATE_FORMAT_SQL))
}

func (s *Sql) SetOrder(order... string) {
	s.order = order
}

func (s *Sql) SetLimit(limit string) {
	s.limit = limit
}

func (s *Sql) GetSQL() (string) {
	query :=
	`SELECT `+strings.Join(s.columns, ",")+"\n" +
	`FROM `+s.from+"\n"
	if len(s.ljoin) > 0 {
		query += "LEFT JOIN "+strings.Join(s.ljoin, " LEFT JOIN ") + "\n"
	}
	query += `WHERE `+ strings.Join(append([]string{"1=1"}, s.where...), " AND ") + "\n"

	if len(s.order) > 0 {
		query += "ORDER BY "+strings.Join(s.order, ",")
	}

	if s.limit != "" {
		query += "LIMIT "+s.limit
	}

	return query
}

func (s *Sql) GetValues() []interface{} {
	return s.values
}

func stringsToInterfaces(values []string)[]interface{}{
	converted := make([]interface{}, len(values))
	for k, v := range values {
		converted[k] = v
	}
	return converted
}

func intsToInterfaces(values []int)[]interface{}{
	converted := make([]interface{}, len(values))
	for k, v := range values {
		converted[k] = v
	}
	return converted
}

func IfOrInt(val bool, trueVal, falseVal int) int {
	if val {
		return trueVal
	}
	return falseVal
}
