//go:generate go tool go-enum -f currency.go
package currency

// ENUM(gbp, eur, usd)
type Symbol string
