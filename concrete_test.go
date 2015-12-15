package main

import "testing"

func TestConcreteSimple(t *testing.T) {
	const input = `
package temperature
import "fmt"
type Celsius float64
func (c Celsius) String() string  { return fmt.Sprintf("%g°C", c) }
func (c *Celsius) SetF(f float64) { *c = Celsius(f - 32 / 9 * 5) }

type Rdr interface {
	Get() (string, error)
	Set(string) error
}
	`

	interfaceName := "Rdr"
	concreteType := "My" + interfaceName
	pkgName := "temperature"
	parseAndPrint(input, interfaceName, concreteType, pkgName)
}
