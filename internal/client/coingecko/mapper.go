package coingecko

func NormalizeSymbol(id string) string {

	switch id {

	case "bitcoin":
		return "BTC"

	case "ethereum":
		return "ETH"

	default:
		return id
	}
}