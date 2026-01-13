package ip

import "strings"

type location struct {
	Name      string
	Acronym   string
	Child     []*location
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}

var (
	center = []location{
		{
			Name:      "Afghanistan",
			Acronym:   "AF",
			Latitude:  "34.5205",
			Longitude: "69.1778",
			Child:     []*location{},
		},
		{
			Name:      "Albania",
			Acronym:   "AL",
			Latitude:  "41.3317",
			Longitude: "19.8318",
			Child:     []*location{},
		},
		{
			Name:      "Algeria",
			Acronym:   "DZ",
			Latitude:  "36.7372",
			Longitude: "3.0865",
			Child:     []*location{},
		},
		{
			Name:      "Andorra",
			Acronym:   "AD",
			Latitude:  "42.5462",
			Longitude: "1.6016",
			Child:     []*location{},
		},
		{
			Name:      "Angola",
			Acronym:   "AO",
			Latitude:  "-8.8383",
			Longitude: "13.2344",
			Child:     []*location{},
		},
		{
			Name:      "Antigua and Barbuda",
			Acronym:   "AG",
			Latitude:  "17.0608",
			Longitude: "-61.7964",
			Child:     []*location{},
		},
		{
			Name:      "Argentina",
			Acronym:   "AR",
			Latitude:  "-34.6118",
			Longitude: "-58.3960",
			Child:     []*location{},
		},
		{
			Name:      "Armenia",
			Acronym:   "AM",
			Latitude:  "40.1792",
			Longitude: "44.4991",
			Child:     []*location{},
		},
		{
			Name:      "Australia",
			Acronym:   "AU",
			Latitude:  "-35.2809",
			Longitude: "149.1300",
			Child: []*location{
				{
					Name:      "New South Wales",
					Acronym:   "NSW",
					Latitude:  "-33.8688",
					Longitude: "151.2093",
					Child:     []*location{},
				},
				{
					Name:      "Victoria",
					Acronym:   "VIC",
					Latitude:  "-37.8136",
					Longitude: "144.9631",
					Child:     []*location{},
				},
				{
					Name:      "Queensland",
					Acronym:   "QLD",
					Latitude:  "-27.4698",
					Longitude: "153.0251",
					Child:     []*location{},
				},
				{
					Name:      "Western Australia",
					Acronym:   "WA",
					Latitude:  "-31.9505",
					Longitude: "115.8605",
					Child:     []*location{},
				},
				{
					Name:      "South Australia",
					Acronym:   "SA",
					Latitude:  "-34.9285",
					Longitude: "138.6007",
					Child:     []*location{},
				},
				{
					Name:      "Tasmania",
					Acronym:   "TAS",
					Latitude:  "-42.8821",
					Longitude: "147.3272",
					Child:     []*location{},
				},
				{
					Name:      "Northern Territory",
					Acronym:   "NT",
					Latitude:  "-12.4634",
					Longitude: "130.8456",
					Child:     []*location{},
				},
				{
					Name:      "Australian Capital Territory",
					Acronym:   "ACT",
					Latitude:  "-35.2809",
					Longitude: "149.1300",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Austria",
			Acronym:   "AT",
			Latitude:  "48.2082",
			Longitude: "16.3738",
			Child:     []*location{},
		},
		{
			Name:      "Azerbaijan",
			Acronym:   "AZ",
			Latitude:  "40.4093",
			Longitude: "49.8671",
			Child:     []*location{},
		},
		{
			Name:      "Bahamas",
			Acronym:   "BS",
			Latitude:  "25.0480",
			Longitude: "-77.3554",
			Child:     []*location{},
		},
		{
			Name:      "Bahrain",
			Acronym:   "BH",
			Latitude:  "26.0667",
			Longitude: "50.5577",
			Child:     []*location{},
		},
		{
			Name:      "Bangladesh",
			Acronym:   "BD",
			Latitude:  "23.8103",
			Longitude: "90.4125",
			Child:     []*location{},
		},
		{
			Name:      "Barbados",
			Acronym:   "BB",
			Latitude:  "13.1132",
			Longitude: "-59.5988",
			Child:     []*location{},
		},
		{
			Name:      "Belarus",
			Acronym:   "BY",
			Latitude:  "53.9045",
			Longitude: "27.5615",
			Child:     []*location{},
		},
		{
			Name:      "Belgium",
			Acronym:   "BE",
			Latitude:  "50.8503",
			Longitude: "4.3517",
			Child:     []*location{},
		},
		{
			Name:      "Belize",
			Acronym:   "BZ",
			Latitude:  "17.2500",
			Longitude: "-88.7667",
			Child:     []*location{},
		},
		{
			Name:      "Benin",
			Acronym:   "BJ",
			Latitude:  "6.3977",
			Longitude: "2.4167",
			Child:     []*location{},
		},
		{
			Name:      "Bhutan",
			Acronym:   "BT",
			Latitude:  "27.4661",
			Longitude: "89.6419",
			Child:     []*location{},
		},
		{
			Name:      "Bolivia",
			Acronym:   "BO",
			Latitude:  "-16.5000",
			Longitude: "-68.1500",
			Child:     []*location{},
		},
		{
			Name:      "Bosnia and Herzegovina",
			Acronym:   "BA",
			Latitude:  "43.8563",
			Longitude: "18.4131",
			Child:     []*location{},
		},
		{
			Name:      "Botswana",
			Acronym:   "BW",
			Latitude:  "-24.6570",
			Longitude: "25.9089",
			Child:     []*location{},
		},
		{
			Name:      "Brazil",
			Acronym:   "BR",
			Latitude:  "-15.8267",
			Longitude: "-47.9218",
			Child: []*location{
				{
					Name:      "São Paulo",
					Acronym:   "SP",
					Latitude:  "-23.5505",
					Longitude: "-46.6333",
					Child:     []*location{},
				},
				{
					Name:      "Rio de Janeiro",
					Acronym:   "RJ",
					Latitude:  "-22.9068",
					Longitude: "-43.1729",
					Child:     []*location{},
				},
				{
					Name:      "Minas Gerais",
					Acronym:   "MG",
					Latitude:  "-19.9167",
					Longitude: "-43.9345",
					Child:     []*location{},
				},
				{
					Name:      "Bahia",
					Acronym:   "BA",
					Latitude:  "-12.9714",
					Longitude: "-38.5014",
					Child:     []*location{},
				},
				{
					Name:      "Rio Grande do Sul",
					Acronym:   "RS",
					Latitude:  "-30.0346",
					Longitude: "-51.2177",
					Child:     []*location{},
				},
				{
					Name:      "Paraná",
					Acronym:   "PR",
					Latitude:  "-25.4284",
					Longitude: "-49.2733",
					Child:     []*location{},
				},
				{
					Name:      "Pernambuco",
					Acronym:   "PE",
					Latitude:  "-8.0476",
					Longitude: "-34.8770",
					Child:     []*location{},
				},
				{
					Name:      "Ceará",
					Acronym:   "CE",
					Latitude:  "-3.7172",
					Longitude: "-38.5434",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Brunei",
			Acronym:   "BN",
			Latitude:  "4.9031",
			Longitude: "114.9398",
			Child:     []*location{},
		},
		{
			Name:      "Bulgaria",
			Acronym:   "BG",
			Latitude:  "42.6977",
			Longitude: "23.3219",
			Child:     []*location{},
		},
		{
			Name:      "Burkina Faso",
			Acronym:   "BF",
			Latitude:  "12.3714",
			Longitude: "-1.5197",
			Child:     []*location{},
		},
		{
			Name:      "Burundi",
			Acronym:   "BI",
			Latitude:  "-3.3842",
			Longitude: "29.3611",
			Child:     []*location{},
		},
		{
			Name:      "Cabo Verde",
			Acronym:   "CV",
			Latitude:  "14.9331",
			Longitude: "-23.5133",
			Child:     []*location{},
		},
		{
			Name:      "Cambodia",
			Acronym:   "KH",
			Latitude:  "11.5564",
			Longitude: "104.9282",
			Child:     []*location{},
		},
		{
			Name:      "Cameroon",
			Acronym:   "CM",
			Latitude:  "3.8480",
			Longitude: "11.5021",
			Child:     []*location{},
		},
		{
			Name:      "Canada",
			Acronym:   "CA",
			Latitude:  "45.4215",
			Longitude: "-75.6972",
			Child: []*location{
				{
					Name:      "Ontario",
					Acronym:   "ON",
					Latitude:  "51.2538",
					Longitude: "-85.3232",
					Child:     []*location{},
				},
				{
					Name:      "Quebec",
					Acronym:   "QC",
					Latitude:  "52.7395",
					Longitude: "-73.4980",
					Child:     []*location{},
				},
				{
					Name:      "British Columbia",
					Acronym:   "BC",
					Latitude:  "53.9333",
					Longitude: "-125.7833",
					Child:     []*location{},
				},
				{
					Name:      "Alberta",
					Acronym:   "AB",
					Latitude:  "53.9333",
					Longitude: "-116.5765",
					Child:     []*location{},
				},
				{
					Name:      "Manitoba",
					Acronym:   "MB",
					Latitude:  "49.8943",
					Longitude: "-97.1385",
					Child:     []*location{},
				},
				{
					Name:      "Saskatchewan",
					Acronym:   "SK",
					Latitude:  "50.8485",
					Longitude: "-106.4520",
					Child:     []*location{},
				},
				{
					Name:      "Nova Scotia",
					Acronym:   "NS",
					Latitude:  "44.6820",
					Longitude: "-63.7443",
					Child:     []*location{},
				},
				{
					Name:      "New Brunswick",
					Acronym:   "NB",
					Latitude:  "46.4983",
					Longitude: "-66.0633",
					Child:     []*location{},
				},
				{
					Name:      "Newfoundland and Labrador",
					Acronym:   "NL",
					Latitude:  "47.5615",
					Longitude: "-52.7126",
					Child:     []*location{},
				},
				{
					Name:      "Prince Edward Island",
					Acronym:   "PE",
					Latitude:  "46.2500",
					Longitude: "-63.0000",
					Child:     []*location{},
				},
				{
					Name:      "Northwest Territories",
					Acronym:   "NT",
					Latitude:  "62.4540",
					Longitude: "-114.3718",
					Child:     []*location{},
				},
				{
					Name:      "Yukon",
					Acronym:   "YT",
					Latitude:  "64.2823",
					Longitude: "-135.0000",
					Child:     []*location{},
				},
				{
					Name:      "Nunavut",
					Acronym:   "NU",
					Latitude:  "70.0000",
					Longitude: "-95.0000",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Central African Republic",
			Acronym:   "CF",
			Latitude:  "4.3947",
			Longitude: "18.5582",
			Child:     []*location{},
		},
		{
			Name:      "Chad",
			Acronym:   "TD",
			Latitude:  "12.1348",
			Longitude: "15.0557",
			Child:     []*location{},
		},
		{
			Name:      "Chile",
			Acronym:   "CL",
			Latitude:  "-33.4489",
			Longitude: "-70.6693",
			Child:     []*location{},
		},
		{
			Name:      "China",
			Acronym:   "CN",
			Latitude:  "39.9042",
			Longitude: "116.4074",
			Child: []*location{
				{
					Name:      "Beijing",
					Acronym:   "BJ",
					Latitude:  "39.9042",
					Longitude: "116.4074",
					Child:     []*location{},
				},
				{
					Name:      "Shanghai",
					Acronym:   "SH",
					Latitude:  "31.2304",
					Longitude: "121.4737",
					Child:     []*location{},
				},
				{
					Name:      "Guangdong",
					Acronym:   "GD",
					Latitude:  "23.1291",
					Longitude: "113.2644",
					Child:     []*location{},
				},
				{
					Name:      "Zhejiang",
					Acronym:   "ZJ",
					Latitude:  "30.2741",
					Longitude: "120.1551",
					Child:     []*location{},
				},
				{
					Name:      "Jiangsu",
					Acronym:   "JS",
					Latitude:  "32.0617",
					Longitude: "118.7778",
					Child:     []*location{},
				},
				{
					Name:      "Sichuan",
					Acronym:   "SC",
					Latitude:  "30.5728",
					Longitude: "104.0668",
					Child:     []*location{},
				},
				{
					Name:      "Hubei",
					Acronym:   "HB",
					Latitude:  "30.5928",
					Longitude: "114.3055",
					Child:     []*location{},
				},
				{
					Name:      "Hunan",
					Acronym:   "HN",
					Latitude:  "28.2282",
					Longitude: "112.9388",
					Child:     []*location{},
				},
				{
					Name:      "Henan",
					Acronym:   "HA",
					Latitude:  "34.7466",
					Longitude: "113.6254",
					Child:     []*location{},
				},
				{
					Name:      "Shandong",
					Acronym:   "SD",
					Latitude:  "36.6758",
					Longitude: "117.0009",
					Child:     []*location{},
				},
				{
					Name:      "Hebei",
					Acronym:   "HE",
					Latitude:  "38.0428",
					Longitude: "114.5149",
					Child:     []*location{},
				},
				{
					Name:      "Liaoning",
					Acronym:   "LN",
					Latitude:  "41.8057",
					Longitude: "123.4315",
					Child:     []*location{},
				},
				{
					Name:      "Jilin",
					Acronym:   "JL",
					Latitude:  "43.8868",
					Longitude: "125.3245",
					Child:     []*location{},
				},
				{
					Name:      "Heilongjiang",
					Acronym:   "HL",
					Latitude:  "45.7732",
					Longitude: "126.6618",
					Child:     []*location{},
				},
				{
					Name:      "Shaanxi",
					Acronym:   "SN",
					Latitude:  "34.3416",
					Longitude: "108.9398",
					Child:     []*location{},
				},
				{
					Name:      "Gansu",
					Acronym:   "GS",
					Latitude:  "36.0611",
					Longitude: "103.8343",
					Child:     []*location{},
				},
				{
					Name:      "Qinghai",
					Acronym:   "QH",
					Latitude:  "36.6171",
					Longitude: "101.7782",
					Child:     []*location{},
				},
				{
					Name:      "Xinjiang",
					Acronym:   "XJ",
					Latitude:  "43.7928",
					Longitude: "87.6177",
					Child:     []*location{},
				},
				{
					Name:      "Tibet",
					Acronym:   "XZ",
					Latitude:  "29.6520",
					Longitude: "91.1720",
					Child:     []*location{},
				},
				{
					Name:      "Guangxi",
					Acronym:   "GX",
					Latitude:  "22.8154",
					Longitude: "108.3275",
					Child:     []*location{},
				},
				{
					Name:      "Inner Mongolia",
					Acronym:   "NM",
					Latitude:  "40.8414",
					Longitude: "111.7519",
					Child:     []*location{},
				},
				{
					Name:      "Ningxia",
					Acronym:   "NX",
					Latitude:  "38.4680",
					Longitude: "106.2731",
					Child:     []*location{},
				},
				{
					Name:      "Hainan",
					Acronym:   "HI",
					Latitude:  "19.2041",
					Longitude: "110.1999",
					Child:     []*location{},
				},
				{
					Name:      "Chongqing",
					Acronym:   "CQ",
					Latitude:  "29.5630",
					Longitude: "106.5516",
					Child:     []*location{},
				},
				{
					Name:      "Tianjin",
					Acronym:   "TJ",
					Latitude:  "39.0842",
					Longitude: "117.2010",
					Child:     []*location{},
				},
				{
					Name:      "Hong Kong",
					Acronym:   "HK",
					Latitude:  "22.3193",
					Longitude: "114.1694",
					Child:     []*location{},
				},
				{
					Name:      "Macao",
					Acronym:   "MO",
					Latitude:  "22.1987",
					Longitude: "113.5439",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Colombia",
			Acronym:   "CO",
			Latitude:  "4.7110",
			Longitude: "-74.0721",
			Child:     []*location{},
		},
		{
			Name:      "Comoros",
			Acronym:   "KM",
			Latitude:  "-11.8750",
			Longitude: "43.3722",
			Child:     []*location{},
		},
		{
			Name:      "Congo",
			Acronym:   "CG",
			Latitude:  "-4.2634",
			Longitude: "15.2429",
			Child:     []*location{},
		},
		{
			Name:      "Congo, Democratic Republic of the",
			Acronym:   "CD",
			Latitude:  "-4.4419",
			Longitude: "15.2663",
			Child:     []*location{},
		},
		{
			Name:      "Costa Rica",
			Acronym:   "CR",
			Latitude:  "9.9333",
			Longitude: "-84.0833",
			Child:     []*location{},
		},
		{
			Name:      "Côte d'Ivoire",
			Acronym:   "CI",
			Latitude:  "5.3600",
			Longitude: "-4.0083",
			Child:     []*location{},
		},
		{
			Name:      "Croatia",
			Acronym:   "HR",
			Latitude:  "45.8150",
			Longitude: "15.9785",
			Child:     []*location{},
		},
		{
			Name:      "Cuba",
			Acronym:   "CU",
			Latitude:  "23.1136",
			Longitude: "-82.3666",
			Child:     []*location{},
		},
		{
			Name:      "Cyprus",
			Acronym:   "CY",
			Latitude:  "35.1856",
			Longitude: "33.3823",
			Child:     []*location{},
		},
		{
			Name:      "Czech Republic",
			Acronym:   "CZ",
			Latitude:  "50.0755",
			Longitude: "14.4378",
			Child:     []*location{},
		},
		{
			Name:      "Denmark",
			Acronym:   "DK",
			Latitude:  "55.6761",
			Longitude: "12.5683",
			Child:     []*location{},
		},
		{
			Name:      "Djibouti",
			Acronym:   "DJ",
			Latitude:  "11.8271",
			Longitude: "42.5905",
			Child:     []*location{},
		},
		{
			Name:      "Dominica",
			Acronym:   "DM",
			Latitude:  "15.4150",
			Longitude: "-61.3710",
			Child:     []*location{},
		},
		{
			Name:      "Dominican Republic",
			Acronym:   "DO",
			Latitude:  "18.4802",
			Longitude: "-69.9381",
			Child:     []*location{},
		},
		{
			Name:      "East Timor",
			Acronym:   "TL",
			Latitude:  "-8.5586",
			Longitude: "125.5736",
			Child:     []*location{},
		},
		{
			Name:      "Ecuador",
			Acronym:   "EC",
			Latitude:  "-0.2295",
			Longitude: "-78.5243",
			Child:     []*location{},
		},
		{
			Name:      "Egypt",
			Acronym:   "EG",
			Latitude:  "30.0444",
			Longitude: "31.2357",
			Child:     []*location{},
		},
		{
			Name:      "El Salvador",
			Acronym:   "SV",
			Latitude:  "13.7021",
			Longitude: "-89.2076",
			Child:     []*location{},
		},
		{
			Name:      "Equatorial Guinea",
			Acronym:   "GQ",
			Latitude:  "3.7452",
			Longitude: "8.7376",
			Child:     []*location{},
		},
		{
			Name:      "Eritrea",
			Acronym:   "ER",
			Latitude:  "15.3333",
			Longitude: "38.9167",
			Child:     []*location{},
		},
		{
			Name:      "Estonia",
			Acronym:   "EE",
			Latitude:  "59.4370",
			Longitude: "24.7536",
			Child:     []*location{},
		},
		{
			Name:      "Eswatini",
			Acronym:   "SZ",
			Latitude:  "-26.3167",
			Longitude: "31.1333",
			Child:     []*location{},
		},
		{
			Name:      "Ethiopia",
			Acronym:   "ET",
			Latitude:  "9.1450",
			Longitude: "40.4897",
			Child:     []*location{},
		},
		{
			Name:      "Fiji",
			Acronym:   "FJ",
			Latitude:  "-18.1248",
			Longitude: "178.4501",
			Child:     []*location{},
		},
		{
			Name:      "Finland",
			Acronym:   "FI",
			Latitude:  "60.1695",
			Longitude: "24.9354",
			Child:     []*location{},
		},
		{
			Name:      "France",
			Acronym:   "FR",
			Latitude:  "48.8566",
			Longitude: "2.3522",
			Child: []*location{
				{
					Name:      "Provence",
					Acronym:   "PAC",
					Latitude:  "43.9352",
					Longitude: "6.0679",
					Child:     []*location{},
				},
				{
					Name:      "Île-de-France",
					Acronym:   "IDF",
					Latitude:  "48.8566",
					Longitude: "2.3522",
					Child:     []*location{},
				},
				{
					Name:      "Nouvelle-Aquitaine",
					Acronym:   "NAQ",
					Latitude:  "44.8378",
					Longitude: "-0.5792",
					Child:     []*location{},
				},
				{
					Name:      "Auvergne",
					Acronym:   "ARA",
					Latitude:  "45.7772",
					Longitude: "3.0870",
					Child:     []*location{},
				},
				{
					Name:      "Occitanie",
					Acronym:   "OCC",
					Latitude:  "43.6047",
					Longitude: "1.4442",
					Child:     []*location{},
				},
				{
					Name:      "Rhône",
					Acronym:   "ARA",
					Latitude:  "45.7640",
					Longitude: "4.8357",
					Child:     []*location{},
				},
				{
					Name:      "Andalusia",
					Acronym:   "AND",
					Latitude:  "37.3891",
					Longitude: "-5.9845",
					Child:     []*location{},
				},
				{
					Name:      "Catalonia",
					Acronym:   "CAT",
					Latitude:  "41.3851",
					Longitude: "2.1734",
					Child:     []*location{},
				},
				{
					Name:      "Madrid",
					Acronym:   "MAD",
					Latitude:  "40.4168",
					Longitude: "-3.7038",
					Child:     []*location{},
				},
				{
					Name:      "Valencia",
					Acronym:   "VAL",
					Latitude:  "39.4699",
					Longitude: "-0.3763",
					Child:     []*location{},
				},
				{
					Name:      "Galicia",
					Acronym:   "GAL",
					Latitude:  "42.6020",
					Longitude: "-8.7076",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Gabon",
			Acronym:   "GA",
			Latitude:  "0.4162",
			Longitude: "9.4476",
			Child:     []*location{},
		},
		{
			Name:      "Gambia",
			Acronym:   "GM",
			Latitude:  "13.4531",
			Longitude: "-16.5775",
			Child:     []*location{},
		},
		{
			Name:      "Georgia",
			Acronym:   "GE",
			Latitude:  "41.7151",
			Longitude: "44.8278",
			Child:     []*location{},
		},
		{
			Name:      "Germany",
			Acronym:   "DE",
			Latitude:  "52.5200",
			Longitude: "13.4050",
			Child: []*location{
				{
					Name:      "Bavaria",
					Acronym:   "BY",
					Latitude:  "48.1351",
					Longitude: "11.5820",
					Child:     []*location{},
				},
				{
					Name:      "North Rhine-Westphalia",
					Acronym:   "NW",
					Latitude:  "51.4332",
					Longitude: "7.6616",
					Child:     []*location{},
				},
				{
					Name:      "Baden-Württemberg",
					Acronym:   "BW",
					Latitude:  "48.7758",
					Longitude: "9.1829",
					Child:     []*location{},
				},
				{
					Name:      "Lower Saxony",
					Acronym:   "NI",
					Latitude:  "52.3759",
					Longitude: "9.7320",
					Child:     []*location{},
				},
				{
					Name:      "Hesse",
					Acronym:   "HE",
					Latitude:  "50.1109",
					Longitude: "8.6821",
					Child:     []*location{},
				},
				{
					Name:      "Saxony",
					Acronym:   "SN",
					Latitude:  "51.0834",
					Longitude: "13.8121",
					Child:     []*location{},
				},
				{
					Name:      "Free State of Bavaria",
					Acronym:   "BY",
					Latitude:  "48.1351",
					Longitude: "11.5820",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Ghana",
			Acronym:   "GH",
			Latitude:  "5.6037",
			Longitude: "-0.1870",
			Child:     []*location{},
		},
		{
			Name:      "Greece",
			Acronym:   "GR",
			Latitude:  "37.9838",
			Longitude: "23.7275",
			Child:     []*location{},
		},
		{
			Name:      "Grenada",
			Acronym:   "GD",
			Latitude:  "12.0528",
			Longitude: "-61.7525",
			Child:     []*location{},
		},
		{
			Name:      "Guatemala",
			Acronym:   "GT",
			Latitude:  "14.6349",
			Longitude: "-90.5069",
			Child:     []*location{},
		},
		{
			Name:      "Guinea",
			Acronym:   "GN",
			Latitude:  "9.6412",
			Longitude: "-13.5784",
			Child:     []*location{},
		},
		{
			Name:      "Guinea-Bissau",
			Acronym:   "GW",
			Latitude:  "11.8594",
			Longitude: "-15.5855",
			Child:     []*location{},
		},
		{
			Name:      "Guyana",
			Acronym:   "GY",
			Latitude:  "6.8045",
			Longitude: "-58.1553",
			Child:     []*location{},
		},
		{
			Name:      "Haiti",
			Acronym:   "HT",
			Latitude:  "18.9714",
			Longitude: "-72.2852",
			Child:     []*location{},
		},
		{
			Name:      "Honduras",
			Acronym:   "HN",
			Latitude:  "14.0723",
			Longitude: "-87.1921",
			Child:     []*location{},
		},
		{
			Name:      "Hungary",
			Acronym:   "HU",
			Latitude:  "47.4979",
			Longitude: "19.0402",
			Child:     []*location{},
		},
		{
			Name:      "Iceland",
			Acronym:   "IS",
			Latitude:  "64.1466",
			Longitude: "-21.9426",
			Child:     []*location{},
		},
		{
			Name:      "India",
			Acronym:   "IN",
			Latitude:  "28.6139",
			Longitude: "77.2090",
			Child: []*location{
				{
					Name:      "Maharashtra",
					Acronym:   "MH",
					Latitude:  "19.0760",
					Longitude: "72.8777",
					Child:     []*location{},
				},
				{
					Name:      "Uttar Pradesh",
					Acronym:   "UP",
					Latitude:  "26.8467",
					Longitude: "80.9462",
					Child:     []*location{},
				},
				{
					Name:      "Bihar",
					Acronym:   "BR",
					Latitude:  "25.5941",
					Longitude: "85.1376",
					Child:     []*location{},
				},
				{
					Name:      "West Bengal",
					Acronym:   "WB",
					Latitude:  "22.5726",
					Longitude: "88.3639",
					Child:     []*location{},
				},
				{
					Name:      "Tamil Nadu",
					Acronym:   "TN",
					Latitude:  "13.0827",
					Longitude: "80.2707",
					Child:     []*location{},
				},
				{
					Name:      "Rajasthan",
					Acronym:   "RJ",
					Latitude:  "26.9124",
					Longitude: "75.7873",
					Child:     []*location{},
				},
				{
					Name:      "Karnataka",
					Acronym:   "KA",
					Latitude:  "12.9716",
					Longitude: "77.5946",
					Child:     []*location{},
				},
				{
					Name:      "Gujarat",
					Acronym:   "GJ",
					Latitude:  "23.2156",
					Longitude: "72.6369",
					Child:     []*location{},
				},
				{
					Name:      "Andhra Pradesh",
					Acronym:   "AP",
					Latitude:  "15.9129",
					Longitude: "79.7400",
					Child:     []*location{},
				},
				{
					Name:      "Odisha",
					Acronym:   "OD",
					Latitude:  "20.2961",
					Longitude: "85.8245",
					Child:     []*location{},
				},
				{
					Name:      "Punjab",
					Acronym:   "PB",
					Latitude:  "31.1471",
					Longitude: "75.3412",
					Child:     []*location{},
				},
				{
					Name:      "Haryana",
					Acronym:   "HR",
					Latitude:  "29.0588",
					Longitude: "76.0856",
					Child:     []*location{},
				},
				{
					Name:      "Kerala",
					Acronym:   "KL",
					Latitude:  "10.8505",
					Longitude: "76.2711",
					Child:     []*location{},
				},
				{
					Name:      "Jharkhand",
					Acronym:   "JH",
					Latitude:  "23.6102",
					Longitude: "85.2799",
					Child:     []*location{},
				},
				{
					Name:      "Assam",
					Acronym:   "AS",
					Latitude:  "26.2006",
					Longitude: "92.9376",
					Child:     []*location{},
				},
				{
					Name:      "Madhya Pradesh",
					Acronym:   "MP",
					Latitude:  "22.9734",
					Longitude: "78.6569",
					Child:     []*location{},
				},
				{
					Name:      "Jammu and Kashmir",
					Acronym:   "JK",
					Latitude:  "34.0837",
					Longitude: "74.7973",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Indonesia",
			Acronym:   "ID",
			Latitude:  "-6.2088",
			Longitude: "106.8456",
			Child:     []*location{},
		},
		{
			Name:      "Iran",
			Acronym:   "IR",
			Latitude:  "35.6892",
			Longitude: "51.3890",
			Child:     []*location{},
		},
		{
			Name:      "Iraq",
			Acronym:   "IQ",
			Latitude:  "33.3152",
			Longitude: "44.3661",
			Child:     []*location{},
		},
		{
			Name:      "Ireland",
			Acronym:   "IE",
			Latitude:  "53.3498",
			Longitude: "-6.2603",
			Child:     []*location{},
		},
		{
			Name:      "Israel",
			Acronym:   "IL",
			Latitude:  "31.7683",
			Longitude: "35.2137",
			Child:     []*location{},
		},
		{
			Name:      "Italy",
			Acronym:   "IT",
			Latitude:  "41.9028",
			Longitude: "12.4964",
			Child: []*location{
				{
					Name:      "Lombardy",
					Acronym:   "LOM",
					Latitude:  "45.4642",
					Longitude: "9.1900",
					Child:     []*location{},
				},
				{
					Name:      "Lazio",
					Acronym:   "LAZ",
					Latitude:  "41.9028",
					Longitude: "12.4964",
					Child:     []*location{},
				},
				{
					Name:      "Sicily",
					Acronym:   "SIC",
					Latitude:  "37.5399",
					Longitude: "15.0805",
					Child:     []*location{},
				},
				{
					Name:      "Campania",
					Acronym:   "CAM",
					Latitude:  "40.8518",
					Longitude: "14.2681",
					Child:     []*location{},
				},
				{
					Name:      "Veneto",
					Acronym:   "VEN",
					Latitude:  "45.4408",
					Longitude: "12.3155",
					Child:     []*location{},
				},
				{
					Name:      "Piedmont",
					Acronym:   "PIE",
					Latitude:  "45.0703",
					Longitude: "7.6869",
					Child:     []*location{},
				},
				{
					Name:      "Apulia",
					Acronym:   "PUG",
					Latitude:  "41.1171",
					Longitude: "16.8719",
					Child:     []*location{},
				},
				{
					Name:      "Tuscany",
					Acronym:   "TOS",
					Latitude:  "43.7711",
					Longitude: "11.2486",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Jamaica",
			Acronym:   "JM",
			Latitude:  "17.9714",
			Longitude: "-76.7931",
			Child:     []*location{},
		},
		{
			Name:      "Japan",
			Acronym:   "JP",
			Latitude:  "35.6762",
			Longitude: "139.6503",
			Child: []*location{
				{
					Name:      "Tokyo",
					Acronym:   "TYO",
					Latitude:  "35.6762",
					Longitude: "139.6503",
					Child:     []*location{},
				},
				{
					Name:      "Osaka",
					Acronym:   "OSK",
					Latitude:  "34.6937",
					Longitude: "135.5023",
					Child:     []*location{},
				},
				{
					Name:      "Kyoto",
					Acronym:   "KYT",
					Latitude:  "35.0116",
					Longitude: "135.7681",
					Child:     []*location{},
				},
				{
					Name:      "Hokkaido",
					Acronym:   "HKD",
					Latitude:  "43.0642",
					Longitude: "141.3469",
					Child:     []*location{},
				},
				{
					Name:      "Okinawa",
					Acronym:   "OKN",
					Latitude:  "26.2124",
					Longitude: "127.6792",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Jordan",
			Acronym:   "JO",
			Latitude:  "31.9632",
			Longitude: "35.9304",
			Child:     []*location{},
		},
		{
			Name:      "Kazakhstan",
			Acronym:   "KZ",
			Latitude:  "51.1605",
			Longitude: "71.4704",
			Child:     []*location{},
		},
		{
			Name:      "Kenya",
			Acronym:   "KE",
			Latitude:  "-1.2921",
			Longitude: "36.8219",
			Child:     []*location{},
		},
		{
			Name:      "Kiribati",
			Acronym:   "KI",
			Latitude:  "1.3278",
			Longitude: "172.9784",
			Child:     []*location{},
		},
		{
			Name:      "Kuwait",
			Acronym:   "KW",
			Latitude:  "29.3759",
			Longitude: "47.9774",
			Child:     []*location{},
		},
		{
			Name:      "Kyrgyzstan",
			Acronym:   "KG",
			Latitude:  "42.8746",
			Longitude: "74.5698",
			Child:     []*location{},
		},
		{
			Name:      "Laos",
			Acronym:   "LA",
			Latitude:  "17.9757",
			Longitude: "102.6061",
			Child:     []*location{},
		},
		{
			Name:      "Latvia",
			Acronym:   "LV",
			Latitude:  "56.9496",
			Longitude: "24.1052",
			Child:     []*location{},
		},
		{
			Name:      "Lebanon",
			Acronym:   "LB",
			Latitude:  "33.8938",
			Longitude: "35.5018",
			Child:     []*location{},
		},
		{
			Name:      "Lesotho",
			Acronym:   "LS",
			Latitude:  "-29.3632",
			Longitude: "27.5144",
			Child:     []*location{},
		},
		{
			Name:      "Liberia",
			Acronym:   "LR",
			Latitude:  "6.2905",
			Longitude: "-10.7605",
			Child:     []*location{},
		},
		{
			Name:      "Libya",
			Acronym:   "LY",
			Latitude:  "32.8872",
			Longitude: "13.1913",
			Child:     []*location{},
		},
		{
			Name:      "Liechtenstein",
			Acronym:   "LI",
			Latitude:  "47.1410",
			Longitude: "9.5215",
			Child:     []*location{},
		},
		{
			Name:      "Lithuania",
			Acronym:   "LT",
			Latitude:  "54.6892",
			Longitude: "25.2797",
			Child:     []*location{},
		},
		{
			Name:      "Luxembourg",
			Acronym:   "LU",
			Latitude:  "49.6116",
			Longitude: "6.1319",
			Child:     []*location{},
		},
		{
			Name:      "Madagascar",
			Acronym:   "MG",
			Latitude:  "-18.8792",
			Longitude: "47.5079",
			Child:     []*location{},
		},
		{
			Name:      "Malawi",
			Acronym:   "MW",
			Latitude:  "-14.0167",
			Longitude: "33.2500",
			Child:     []*location{},
		},
		{
			Name:      "Malaysia",
			Acronym:   "MY",
			Latitude:  "3.1390",
			Longitude: "101.6869",
			Child:     []*location{},
		},
		{
			Name:      "Maldives",
			Acronym:   "MV",
			Latitude:  "4.1755",
			Longitude: "73.5093",
			Child:     []*location{},
		},
		{
			Name:      "Mali",
			Acronym:   "ML",
			Latitude:  "12.6392",
			Longitude: "-8.0029",
			Child:     []*location{},
		},
		{
			Name:      "Malta",
			Acronym:   "MT",
			Latitude:  "35.8989",
			Longitude: "14.5146",
			Child:     []*location{},
		},
		{
			Name:      "Marshall Islands",
			Acronym:   "MH",
			Latitude:  "7.1315",
			Longitude: "171.1845",
			Child:     []*location{},
		},
		{
			Name:      "Mauritania",
			Acronym:   "MR",
			Latitude:  "18.0735",
			Longitude: "-15.9582",
			Child:     []*location{},
		},
		{
			Name:      "Mauritius",
			Acronym:   "MU",
			Latitude:  "-20.2675",
			Longitude: "57.5271",
			Child:     []*location{},
		},
		{
			Name:      "Mexico",
			Acronym:   "MX",
			Latitude:  "19.4326",
			Longitude: "-99.1332",
			Child:     []*location{},
		},
		{
			Name:      "Micronesia",
			Acronym:   "FM",
			Latitude:  "6.9233",
			Longitude: "158.1610",
			Child:     []*location{},
		},
		{
			Name:      "Moldova",
			Acronym:   "MD",
			Latitude:  "47.0105",
			Longitude: "28.8638",
			Child:     []*location{},
		},
		{
			Name:      "Monaco",
			Acronym:   "MC",
			Latitude:  "43.7347",
			Longitude: "7.4206",
			Child:     []*location{},
		},
		{
			Name:      "Mongolia",
			Acronym:   "MN",
			Latitude:  "47.9203",
			Longitude: "106.9172",
			Child:     []*location{},
		},
		{
			Name:      "Montenegro",
			Acronym:   "ME",
			Latitude:  "42.4413",
			Longitude: "19.2629",
			Child:     []*location{},
		},
		{
			Name:      "Morocco",
			Acronym:   "MA",
			Latitude:  "33.9716",
			Longitude: "-6.8498",
			Child:     []*location{},
		},
		{
			Name:      "Mozambique",
			Acronym:   "MZ",
			Latitude:  "-25.9692",
			Longitude: "32.5832",
			Child:     []*location{},
		},
		{
			Name:      "Myanmar",
			Acronym:   "MM",
			Latitude:  "19.7633",
			Longitude: "96.0785",
			Child:     []*location{},
		},
		{
			Name:      "Namibia",
			Acronym:   "NA",
			Latitude:  "-22.5609",
			Longitude: "17.0658",
			Child:     []*location{},
		},
		{
			Name:      "Nauru",
			Acronym:   "NR",
			Latitude:  "-0.5478",
			Longitude: "166.9313",
			Child:     []*location{},
		},
		{
			Name:      "Nepal",
			Acronym:   "NP",
			Latitude:  "27.7172",
			Longitude: "85.3240",
			Child:     []*location{},
		},
		{
			Name:      "Netherlands",
			Acronym:   "NL",
			Latitude:  "52.3676",
			Longitude: "4.9041",
			Child:     []*location{},
		},
		{
			Name:      "New Zealand",
			Acronym:   "NZ",
			Latitude:  "-41.2865",
			Longitude: "174.7762",
			Child:     []*location{},
		},
		{
			Name:      "Nicaragua",
			Acronym:   "NI",
			Latitude:  "12.1144",
			Longitude: "-86.2362",
			Child:     []*location{},
		},
		{
			Name:      "Niger",
			Acronym:   "NE",
			Latitude:  "13.5116",
			Longitude: "2.1254",
			Child:     []*location{},
		},
		{
			Name:      "Nigeria",
			Acronym:   "NG",
			Latitude:  "6.5244",
			Longitude: "3.3792",
			Child:     []*location{},
		},
		{
			Name:      "North Macedonia",
			Acronym:   "MK",
			Latitude:  "41.9981",
			Longitude: "21.4254",
			Child:     []*location{},
		},
		{
			Name:      "Norway",
			Acronym:   "NO",
			Latitude:  "59.9139",
			Longitude: "10.7522",
			Child:     []*location{},
		},
		{
			Name:      "Oman",
			Acronym:   "OM",
			Latitude:  "23.5859",
			Longitude: "58.4059",
			Child:     []*location{},
		},
		{
			Name:      "Pakistan",
			Acronym:   "PK",
			Latitude:  "33.6844",
			Longitude: "73.0479",
			Child:     []*location{},
		},
		{
			Name:      "Palau",
			Acronym:   "PW",
			Latitude:  "7.51498",
			Longitude: "134.58252",
			Child:     []*location{},
		},
		{
			Name:      "Panama",
			Acronym:   "PA",
			Latitude:  "8.9824",
			Longitude: "-79.5199",
			Child:     []*location{},
		},
		{
			Name:      "Papua New Guinea",
			Acronym:   "PG",
			Latitude:  "-9.4438",
			Longitude: "147.1803",
			Child:     []*location{},
		},
		{
			Name:      "Paraguay",
			Acronym:   "PY",
			Latitude:  "-25.2637",
			Longitude: "-57.5759",
			Child:     []*location{},
		},
		{
			Name:      "Peru",
			Acronym:   "PE",
			Latitude:  "-12.0464",
			Longitude: "-77.0428",
			Child:     []*location{},
		},
		{
			Name:      "Philippines",
			Acronym:   "PH",
			Latitude:  "14.5995",
			Longitude: "120.9842",
			Child:     []*location{},
		},
		{
			Name:      "Poland",
			Acronym:   "PL",
			Latitude:  "52.2297",
			Longitude: "21.0122",
			Child:     []*location{},
		},
		{
			Name:      "Portugal",
			Acronym:   "PT",
			Latitude:  "38.7223",
			Longitude: "-9.1393",
			Child:     []*location{},
		},
		{
			Name:      "Qatar",
			Acronym:   "QA",
			Latitude:  "25.2854",
			Longitude: "51.5310",
			Child:     []*location{},
		},
		{
			Name:      "Romania",
			Acronym:   "RO",
			Latitude:  "44.4268",
			Longitude: "26.1025",
			Child:     []*location{},
		},
		{
			Name:      "Russia",
			Acronym:   "RU",
			Latitude:  "55.7558",
			Longitude: "37.6173",
			Child:     []*location{},
		},
		{
			Name:      "Rwanda",
			Acronym:   "RW",
			Latitude:  "-1.9403",
			Longitude: "29.8739",
			Child:     []*location{},
		},
		{
			Name:      "Saint Kitts and Nevis",
			Acronym:   "KN",
			Latitude:  "17.2923",
			Longitude: "-62.7325",
			Child:     []*location{},
		},
		{
			Name:      "Saint Lucia",
			Acronym:   "LC",
			Latitude:  "13.9074",
			Longitude: "-60.9789",
			Child:     []*location{},
		},
		{
			Name:      "Saint Vincent and the Grenadines",
			Acronym:   "VC",
			Latitude:  "13.1605",
			Longitude: "-61.2226",
			Child:     []*location{},
		},
		{
			Name:      "Samoa",
			Acronym:   "WS",
			Latitude:  "-13.8506",
			Longitude: "-171.7513",
			Child:     []*location{},
		},
		{
			Name:      "San Marino",
			Acronym:   "SM",
			Latitude:  "43.9424",
			Longitude: "12.4578",
			Child:     []*location{},
		},
		{
			Name:      "São Tomé and Príncipe",
			Acronym:   "ST",
			Latitude:  "0.3364",
			Longitude: "6.7304",
			Child:     []*location{},
		},
		{
			Name:      "Saudi Arabia",
			Acronym:   "SA",
			Latitude:  "24.7136",
			Longitude: "46.6753",
			Child:     []*location{},
		},
		{
			Name:      "Senegal",
			Acronym:   "SN",
			Latitude:  "14.7167",
			Longitude: "-17.4677",
			Child:     []*location{},
		},
		{
			Name:      "Serbia",
			Acronym:   "RS",
			Latitude:  "44.7866",
			Longitude: "20.4489",
			Child:     []*location{},
		},
		{
			Name:      "Seychelles",
			Acronym:   "SC",
			Latitude:  "-4.6796",
			Longitude: "55.4920",
			Child:     []*location{},
		},
		{
			Name:      "Sierra Leone",
			Acronym:   "SL",
			Latitude:  "8.4657",
			Longitude: "-13.2317",
			Child:     []*location{},
		},
		{
			Name:      "Singapore",
			Acronym:   "SG",
			Latitude:  "1.3521",
			Longitude: "103.8198",
			Child:     []*location{},
		},
		{
			Name:      "Slovakia",
			Acronym:   "SK",
			Latitude:  "48.1476",
			Longitude: "17.1077",
			Child:     []*location{},
		},
		{
			Name:      "Slovenia",
			Acronym:   "SI",
			Latitude:  "46.0569",
			Longitude: "14.5058",
			Child:     []*location{},
		},
		{
			Name:      "Solomon Islands",
			Acronym:   "SB",
			Latitude:  "-9.4456",
			Longitude: "159.9728",
			Child:     []*location{},
		},
		{
			Name:      "Somalia",
			Acronym:   "SO",
			Latitude:  "2.0469",
			Longitude: "45.3182",
			Child:     []*location{},
		},
		{
			Name:      "South Africa",
			Acronym:   "ZA",
			Latitude:  "-25.7479",
			Longitude: "28.2293",
			Child:     []*location{},
		},
		{
			Name:      "South Korea",
			Acronym:   "KR",
			Latitude:  "37.5665",
			Longitude: "126.9780",
			Child: []*location{
				{
					Name:      "Seoul",
					Acronym:   "SEO",
					Latitude:  "37.5665",
					Longitude: "126.9780",
					Child:     []*location{},
				},
				{
					Name:      "Busan",
					Acronym:   "BUS",
					Latitude:  "35.1796",
					Longitude: "129.0756",
					Child:     []*location{},
				},
				{
					Name:      "Daegu",
					Acronym:   "DAE",
					Latitude:  "35.8722",
					Longitude: "128.6025",
					Child:     []*location{},
				},
				{
					Name:      "Incheon",
					Acronym:   "INC",
					Latitude:  "37.4563",
					Longitude: "126.7052",
					Child:     []*location{},
				},
				{
					Name:      "Gwangju",
					Acronym:   "GWJ",
					Latitude:  "35.1595",
					Longitude: "126.8526",
					Child:     []*location{},
				},
				{
					Name:      "Daejeon",
					Acronym:   "DAE",
					Latitude:  "36.3504",
					Longitude: "127.3845",
					Child:     []*location{},
				},
				{
					Name:      "Ulsan",
					Acronym:   "ULS",
					Latitude:  "35.5384",
					Longitude: "129.3114",
					Child:     []*location{},
				},
				{
					Name:      "Gyeonggi Province",
					Acronym:   "GG",
					Latitude:  "37.2752",
					Longitude: "127.0095",
					Child:     []*location{},
				},
				{
					Name:      "Gangwon Province",
					Acronym:   "GW",
					Latitude:  "37.5558",
					Longitude: "128.2093",
					Child:     []*location{},
				},
				{
					Name:      "North Chungcheong Province",
					Acronym:   "NC",
					Latitude:  "36.6284",
					Longitude: "127.9288",
					Child:     []*location{},
				},
				{
					Name:      "South Chungcheong Province",
					Acronym:   "SC",
					Latitude:  "36.5581",
					Longitude: "126.7990",
					Child:     []*location{},
				},
				{
					Name:      "North Jeolla Province",
					Acronym:   "NJ",
					Latitude:  "35.7175",
					Longitude: "127.1539",
					Child:     []*location{},
				},
				{
					Name:      "South Jeolla Province",
					Acronym:   "SJ",
					Latitude:  "34.8160",
					Longitude: "126.9218",
					Child:     []*location{},
				},
				{
					Name:      "North Gyeongsang Province",
					Acronym:   "NG",
					Latitude:  "36.4919",
					Longitude: "128.8889",
					Child:     []*location{},
				},
				{
					Name:      "South Gyeongsang Province",
					Acronym:   "SG",
					Latitude:  "35.2380",
					Longitude: "128.6919",
					Child:     []*location{},
				},
				{
					Name:      "Jeju Province",
					Acronym:   "JJ",
					Latitude:  "33.4996",
					Longitude: "126.5312",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "South Sudan",
			Acronym:   "SS",
			Latitude:  "4.8594",
			Longitude: "31.5713",
			Child:     []*location{},
		},
		{
			Name:      "Spain",
			Acronym:   "ES",
			Latitude:  "40.4168",
			Longitude: "-3.7038",
			Child:     []*location{},
		},
		{
			Name:      "Sri Lanka",
			Acronym:   "LK",
			Latitude:  "6.9271",
			Longitude: "79.8612",
			Child:     []*location{},
		},
		{
			Name:      "Sudan",
			Acronym:   "SD",
			Latitude:  "15.5007",
			Longitude: "32.5599",
			Child:     []*location{},
		},
		{
			Name:      "Suriname",
			Acronym:   "SR",
			Latitude:  "5.8520",
			Longitude: "-55.2038",
			Child:     []*location{},
		},
		{
			Name:      "Sweden",
			Acronym:   "SE",
			Latitude:  "59.3293",
			Longitude: "18.0686",
			Child:     []*location{},
		},
		{
			Name:      "Switzerland",
			Acronym:   "CH",
			Latitude:  "46.9481",
			Longitude: "7.4474",
			Child:     []*location{},
		},
		{
			Name:      "Syria",
			Acronym:   "SY",
			Latitude:  "33.5138",
			Longitude: "36.2765",
			Child:     []*location{},
		},
		{
			Name:      "Taiwan",
			Acronym:   "TW",
			Latitude:  "25.0330",
			Longitude: "121.5654",
			Child:     []*location{},
		},
		{
			Name:      "Tajikistan",
			Acronym:   "TJ",
			Latitude:  "38.5598",
			Longitude: "68.7864",
			Child:     []*location{},
		},
		{
			Name:      "Tanzania",
			Acronym:   "TZ",
			Latitude:  "-6.7924",
			Longitude: "39.2083",
			Child:     []*location{},
		},
		{
			Name:      "Thailand",
			Acronym:   "TH",
			Latitude:  "13.7563",
			Longitude: "100.5018",
			Child:     []*location{},
		},
		{
			Name:      "Togo",
			Acronym:   "TG",
			Latitude:  "6.1395",
			Longitude: "1.2255",
			Child:     []*location{},
		},
		{
			Name:      "Tonga",
			Acronym:   "TO",
			Latitude:  "-21.1789",
			Longitude: "-175.1982",
			Child:     []*location{},
		},
		{
			Name:      "Trinidad and Tobago",
			Acronym:   "TT",
			Latitude:  "10.6918",
			Longitude: "-61.2225",
			Child:     []*location{},
		},
		{
			Name:      "Tunisia",
			Acronym:   "TN",
			Latitude:  "36.8065",
			Longitude: "10.1815",
			Child:     []*location{},
		},
		{
			Name:      "Turkey",
			Acronym:   "TR",
			Latitude:  "39.9334",
			Longitude: "32.8597",
			Child:     []*location{},
		},
		{
			Name:      "Turkmenistan",
			Acronym:   "TM",
			Latitude:  "37.9601",
			Longitude: "58.3261",
			Child:     []*location{},
		},
		{
			Name:      "Tuvalu",
			Acronym:   "TV",
			Latitude:  "-8.5244",
			Longitude: "179.1906",
			Child:     []*location{},
		},
		{
			Name:      "Uganda",
			Acronym:   "UG",
			Latitude:  "0.3476",
			Longitude: "32.5825",
			Child:     []*location{},
		},
		{
			Name:      "Ukraine",
			Acronym:   "UA",
			Latitude:  "50.4501",
			Longitude: "30.5234",
			Child:     []*location{},
		},
		{
			Name:      "United Arab Emirates",
			Acronym:   "AE",
			Latitude:  "25.2769",
			Longitude: "55.2962",
			Child:     []*location{},
		},
		{
			Name:      "United Kingdom",
			Acronym:   "GB",
			Latitude:  "51.5074",
			Longitude: "-0.1278",
			Child: []*location{
				{
					Name:      "England",
					Acronym:   "ENG",
					Latitude:  "52.3555",
					Longitude: "-1.1743",
					Child:     []*location{},
				},
				{
					Name:      "Scotland",
					Acronym:   "SCT",
					Latitude:  "56.4907",
					Longitude: "-4.2026",
					Child:     []*location{},
				},
				{
					Name:      "Wales",
					Acronym:   "WLS",
					Latitude:  "52.1307",
					Longitude: "-3.7837",
					Child:     []*location{},
				},
				{
					Name:      "Northern Ireland",
					Acronym:   "NIR",
					Latitude:  "54.7877",
					Longitude: "-6.4923",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "United States",
			Acronym:   "US",
			Latitude:  "38.9072",
			Longitude: "-77.0369",
			Child: []*location{
				{
					Name:      "California",
					Acronym:   "CA",
					Latitude:  "36.7783",
					Longitude: "-119.4179",
					Child:     []*location{},
				},
				{
					Name:      "New York",
					Acronym:   "NY",
					Latitude:  "43.0000",
					Longitude: "-75.0000",
					Child:     []*location{},
				},
				{
					Name:      "Texas",
					Acronym:   "TX",
					Latitude:  "31.0000",
					Longitude: "-100.0000",
					Child:     []*location{},
				},
				{
					Name:      "Florida",
					Acronym:   "FL",
					Latitude:  "27.7663",
					Longitude: "-82.4754",
					Child:     []*location{},
				},
				{
					Name:      "Hawaii",
					Acronym:   "HI",
					Latitude:  "21.3099",
					Longitude: "-157.8581",
					Child:     []*location{},
				},
				{
					Name:      "Alaska",
					Acronym:   "AK",
					Latitude:  "61.0000",
					Longitude: "-150.0000",
					Child:     []*location{},
				},
				{
					Name:      "Washington",
					Acronym:   "WA",
					Latitude:  "47.7511",
					Longitude: "-120.7401",
					Child:     []*location{},
				},
				{
					Name:      "Massachusetts",
					Acronym:   "MA",
					Latitude:  "42.4072",
					Longitude: "-71.3824",
					Child:     []*location{},
				},
				{
					Name:      "Pennsylvania",
					Acronym:   "PA",
					Latitude:  "41.2033",
					Longitude: "-77.1945",
					Child:     []*location{},
				},
				{
					Name:      "Illinois",
					Acronym:   "IL",
					Latitude:  "40.6331",
					Longitude: "-89.3985",
					Child:     []*location{},
				},
				{
					Name:      "Ohio",
					Acronym:   "OH",
					Latitude:  "40.4173",
					Longitude: "-82.9071",
					Child:     []*location{},
				},
				{
					Name:      "Michigan",
					Acronym:   "MI",
					Latitude:  "44.3148",
					Longitude: "-85.6024",
					Child:     []*location{},
				},
				{
					Name:      "New Jersey",
					Acronym:   "NJ",
					Latitude:  "40.0583",
					Longitude: "-74.4057",
					Child:     []*location{},
				},
				{
					Name:      "Virginia",
					Acronym:   "VA",
					Latitude:  "37.4316",
					Longitude: "-78.6569",
					Child:     []*location{},
				},
				{
					Name:      "North Carolina",
					Acronym:   "NC",
					Latitude:  "35.7596",
					Longitude: "-79.0193",
					Child:     []*location{},
				},
				{
					Name:      "Georgia",
					Acronym:   "GA",
					Latitude:  "32.1656",
					Longitude: "-82.9001",
					Child:     []*location{},
				},
			},
		},
		{
			Name:      "Uruguay",
			Acronym:   "UY",
			Latitude:  "-34.8836",
			Longitude: "-56.1819",
			Child:     []*location{},
		},
		{
			Name:      "Uzbekistan",
			Acronym:   "UZ",
			Latitude:  "41.2995",
			Longitude: "69.2401",
			Child:     []*location{},
		},
		{
			Name:      "Vanuatu",
			Acronym:   "VU",
			Latitude:  "-17.7338",
			Longitude: "168.3215",
			Child:     []*location{},
		},
		{
			Name:      "Vatican City",
			Acronym:   "VA",
			Latitude:  "41.9029",
			Longitude: "12.4534",
			Child:     []*location{},
		},
		{
			Name:      "Venezuela",
			Acronym:   "VE",
			Latitude:  "10.4806",
			Longitude: "-66.9036",
			Child:     []*location{},
		},
		{
			Name:      "Vietnam",
			Acronym:   "VN",
			Latitude:  "21.0285",
			Longitude: "105.8542",
			Child:     []*location{},
		},
		{
			Name:      "Yemen",
			Acronym:   "YE",
			Latitude:  "15.3694",
			Longitude: "44.1910",
			Child:     []*location{},
		},
		{
			Name:      "Zambia",
			Acronym:   "ZM",
			Latitude:  "-15.3875",
			Longitude: "28.3228",
			Child:     []*location{},
		},
		{
			Name:      "Zimbabwe",
			Acronym:   "ZW",
			Latitude:  "-17.8252",
			Longitude: "31.0335",
			Child:     []*location{},
		},
	}
)

// GetCapitalCoordinates retrieves capital coordinates by country name, code, or province/state name
// Supports both single input and combined country+region input (e.g., "CN, Beijing" or "US, California")
func GetCapitalCoordinates(input string) (latitude, longitude string, found bool) {
	if input == "" {
		return "", "", false
	}

	input = strings.TrimSpace(input)

	// Check if input is empty after trimming
	if input == "" {
		return "", "", false
	}

	// Check if input contains comma (combined country+region)
	if strings.Contains(input, ",") {
		return handleCombinedInput(input)
	}

	// Handle single input (existing logic)
	inputLower := strings.ToLower(input)

	// First, search through all provinces/states to prioritize them over countries with same acronyms
	for _, country := range center {
		// Check if input matches any child (province/state) - return country's capital coordinates
		for _, child := range country.Child {
			if strings.ToLower(child.Name) == inputLower || strings.ToLower(child.Acronym) == inputLower {
				return country.Latitude, country.Longitude, true
			}
		}
	}

	// Then search through all countries (only if no province/state match found)
	for _, country := range center {
		// Check exact country name match
		if strings.ToLower(country.Name) == inputLower {
			return country.Latitude, country.Longitude, true
		}

		// Check country acronym match
		if strings.ToLower(country.Acronym) == inputLower {
			return country.Latitude, country.Longitude, true
		}
	}

	// If no exact match, try fuzzy matching for country names
	for _, country := range center {
		if strings.Contains(strings.ToLower(country.Name), inputLower) ||
			strings.Contains(inputLower, strings.ToLower(country.Name)) {
			return country.Latitude, country.Longitude, true
		}
	}

	// If still no match, try fuzzy matching for province/state names
	for _, country := range center {
		for _, child := range country.Child {
			if strings.Contains(strings.ToLower(child.Name), inputLower) ||
				strings.Contains(inputLower, strings.ToLower(child.Name)) {
				return country.Latitude, country.Longitude, true
			}
		}
	}

	return "", "", false
}

// handleCombinedInput processes input with comma separating country and region
func handleCombinedInput(input string) (latitude, longitude string, found bool) {
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return "", "", false
	}

	countryPart := strings.TrimSpace(parts[0])
	regionPart := strings.TrimSpace(parts[1])

	countryLower := strings.ToLower(countryPart)
	regionLower := strings.ToLower(regionPart)

	// First find the matching country
	var matchedCountry *location
	for _, country := range center {
		if strings.ToLower(country.Name) == countryLower ||
			strings.ToLower(country.Acronym) == countryLower {
			matchedCountry = &country
			break
		}
	}

	// If no exact country match, try fuzzy matching
	if matchedCountry == nil {
		for _, country := range center {
			if strings.Contains(strings.ToLower(country.Name), countryLower) ||
				strings.Contains(countryLower, strings.ToLower(country.Name)) {
				matchedCountry = &country
				break
			}
		}
	}

	// If still no country match, try to find the country that contains the province/state matching countryPart
	if matchedCountry == nil {
		for _, country := range center {
			for _, child := range country.Child {
				if strings.ToLower(child.Name) == countryLower ||
					strings.ToLower(child.Acronym) == countryLower {
					matchedCountry = &country
					break
				}
			}
			if matchedCountry != nil {
				break
			}
		}
	}

	// If still no match, try fuzzy matching for provinces/states
	if matchedCountry == nil {
		for _, country := range center {
			for _, child := range country.Child {
				if strings.Contains(strings.ToLower(child.Name), countryLower) ||
					strings.Contains(countryLower, strings.ToLower(child.Name)) {
					matchedCountry = &country
					break
				}
			}
			if matchedCountry != nil {
				break
			}
		}
	}

	// If still no match, return false
	if matchedCountry == nil {
		return "", "", false
	}

	// Check if region matches any child of the found country
	for _, child := range matchedCountry.Child {
		if strings.ToLower(child.Name) == regionLower ||
			strings.ToLower(child.Acronym) == regionLower {
			// Return the country's capital coordinates
			return matchedCountry.Latitude, matchedCountry.Longitude, true
		}
	}

	// If no exact region match, try fuzzy matching within the country's children
	for _, child := range matchedCountry.Child {
		if strings.Contains(strings.ToLower(child.Name), regionLower) ||
			strings.Contains(regionLower, strings.ToLower(child.Name)) {
			return matchedCountry.Latitude, matchedCountry.Longitude, true
		}
	}

	// If region doesn't match but country does, return country's capital coordinates
	return matchedCountry.Latitude, matchedCountry.Longitude, true
}
