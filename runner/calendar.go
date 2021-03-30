package runner

import "fmt"

type Calendar struct {
	MinsSinceEpoch  int
	incSpeed, ticks float64
}

func NewCalendar(mins, hours, day, month, year int, incSpeed float64) *Calendar {
	return &Calendar{toEpoch(mins, hours, day, month, year), incSpeed, 0}
}

func (c *Calendar) Incr(delta float64) {
	c.ticks += delta
	if c.ticks >= c.incSpeed {
		c.ticks = 0
		c.MinsSinceEpoch++
		if c.MinsSinceEpoch%100 == 0 {
			fmt.Printf("%s\n", c.AsString())
		}
	}
}

const (
	ticksPerHour  = 60
	ticksPerDay   = ticksPerHour * 24
	ticksPerMonth = ticksPerDay * 30
	ticksPerYear  = ticksPerMonth * 12
)

func (c *Calendar) FromEpoch() (mins, hours, day, month, year int) {
	year = c.MinsSinceEpoch / ticksPerYear
	month = (c.MinsSinceEpoch % ticksPerYear) / ticksPerMonth
	day = ((c.MinsSinceEpoch % ticksPerYear) % ticksPerMonth) / ticksPerDay
	hours = (((c.MinsSinceEpoch % ticksPerYear) % ticksPerMonth) % ticksPerDay) / ticksPerHour
	mins = (((c.MinsSinceEpoch % ticksPerYear) % ticksPerMonth) % ticksPerDay) % ticksPerHour
	return
}

func (c *Calendar) AsString() string {
	mins, hours, day, month, year := c.FromEpoch()
	return fmt.Sprintf("%d/%d/%d %d:%d", year, month, day, hours, mins)
}

func toEpoch(mins, hours, day, month, year int) int {
	return year*ticksPerYear + month*ticksPerMonth + day*ticksPerDay + hours*ticksPerHour + mins
}
