package pbmoney

import (
	"fmt"
	"math"
	"strings"

	"go.pedge.io/pb/go/google/type"
)

var (
	//CurrencyCodeAED is a simple holder for CurrencyCode_CURRENCY_CODE_AED for simplicity.
	CurrencyCodeAED = CurrencyCode_CURRENCY_CODE_AED
	//CurrencyCodeAFN is a simple holder for CurrencyCode_CURRENCY_CODE_AFN for simplicity.
	CurrencyCodeAFN = CurrencyCode_CURRENCY_CODE_AFN
	//CurrencyCodeALL is a simple holder for CurrencyCode_CURRENCY_CODE_ALL for simplicity.
	CurrencyCodeALL = CurrencyCode_CURRENCY_CODE_ALL
	//CurrencyCodeAMD is a simple holder for CurrencyCode_CURRENCY_CODE_AMD for simplicity.
	CurrencyCodeAMD = CurrencyCode_CURRENCY_CODE_AMD
	//CurrencyCodeANG is a simple holder for CurrencyCode_CURRENCY_CODE_ANG for simplicity.
	CurrencyCodeANG = CurrencyCode_CURRENCY_CODE_ANG
	//CurrencyCodeAOA is a simple holder for CurrencyCode_CURRENCY_CODE_AOA for simplicity.
	CurrencyCodeAOA = CurrencyCode_CURRENCY_CODE_AOA
	//CurrencyCodeARS is a simple holder for CurrencyCode_CURRENCY_CODE_ARS for simplicity.
	CurrencyCodeARS = CurrencyCode_CURRENCY_CODE_ARS
	//CurrencyCodeAUD is a simple holder for CurrencyCode_CURRENCY_CODE_AUD for simplicity.
	CurrencyCodeAUD = CurrencyCode_CURRENCY_CODE_AUD
	//CurrencyCodeAWG is a simple holder for CurrencyCode_CURRENCY_CODE_AWG for simplicity.
	CurrencyCodeAWG = CurrencyCode_CURRENCY_CODE_AWG
	//CurrencyCodeAZN is a simple holder for CurrencyCode_CURRENCY_CODE_AZN for simplicity.
	CurrencyCodeAZN = CurrencyCode_CURRENCY_CODE_AZN
	//CurrencyCodeBAM is a simple holder for CurrencyCode_CURRENCY_CODE_BAM for simplicity.
	CurrencyCodeBAM = CurrencyCode_CURRENCY_CODE_BAM
	//CurrencyCodeBBD is a simple holder for CurrencyCode_CURRENCY_CODE_BBD for simplicity.
	CurrencyCodeBBD = CurrencyCode_CURRENCY_CODE_BBD
	//CurrencyCodeBDT is a simple holder for CurrencyCode_CURRENCY_CODE_BDT for simplicity.
	CurrencyCodeBDT = CurrencyCode_CURRENCY_CODE_BDT
	//CurrencyCodeBGN is a simple holder for CurrencyCode_CURRENCY_CODE_BGN for simplicity.
	CurrencyCodeBGN = CurrencyCode_CURRENCY_CODE_BGN
	//CurrencyCodeBHD is a simple holder for CurrencyCode_CURRENCY_CODE_BHD for simplicity.
	CurrencyCodeBHD = CurrencyCode_CURRENCY_CODE_BHD
	//CurrencyCodeBIF is a simple holder for CurrencyCode_CURRENCY_CODE_BIF for simplicity.
	CurrencyCodeBIF = CurrencyCode_CURRENCY_CODE_BIF
	//CurrencyCodeBMD is a simple holder for CurrencyCode_CURRENCY_CODE_BMD for simplicity.
	CurrencyCodeBMD = CurrencyCode_CURRENCY_CODE_BMD
	//CurrencyCodeBND is a simple holder for CurrencyCode_CURRENCY_CODE_BND for simplicity.
	CurrencyCodeBND = CurrencyCode_CURRENCY_CODE_BND
	//CurrencyCodeBOB is a simple holder for CurrencyCode_CURRENCY_CODE_BOB for simplicity.
	CurrencyCodeBOB = CurrencyCode_CURRENCY_CODE_BOB
	//CurrencyCodeBRL is a simple holder for CurrencyCode_CURRENCY_CODE_BRL for simplicity.
	CurrencyCodeBRL = CurrencyCode_CURRENCY_CODE_BRL
	//CurrencyCodeBSD is a simple holder for CurrencyCode_CURRENCY_CODE_BSD for simplicity.
	CurrencyCodeBSD = CurrencyCode_CURRENCY_CODE_BSD
	//CurrencyCodeBWP is a simple holder for CurrencyCode_CURRENCY_CODE_BWP for simplicity.
	CurrencyCodeBWP = CurrencyCode_CURRENCY_CODE_BWP
	//CurrencyCodeBYR is a simple holder for CurrencyCode_CURRENCY_CODE_BYR for simplicity.
	CurrencyCodeBYR = CurrencyCode_CURRENCY_CODE_BYR
	//CurrencyCodeBZD is a simple holder for CurrencyCode_CURRENCY_CODE_BZD for simplicity.
	CurrencyCodeBZD = CurrencyCode_CURRENCY_CODE_BZD
	//CurrencyCodeCAD is a simple holder for CurrencyCode_CURRENCY_CODE_CAD for simplicity.
	CurrencyCodeCAD = CurrencyCode_CURRENCY_CODE_CAD
	//CurrencyCodeCHF is a simple holder for CurrencyCode_CURRENCY_CODE_CHF for simplicity.
	CurrencyCodeCHF = CurrencyCode_CURRENCY_CODE_CHF
	//CurrencyCodeCLP is a simple holder for CurrencyCode_CURRENCY_CODE_CLP for simplicity.
	CurrencyCodeCLP = CurrencyCode_CURRENCY_CODE_CLP
	//CurrencyCodeCNY is a simple holder for CurrencyCode_CURRENCY_CODE_CNY for simplicity.
	CurrencyCodeCNY = CurrencyCode_CURRENCY_CODE_CNY
	//CurrencyCodeCOP is a simple holder for CurrencyCode_CURRENCY_CODE_COP for simplicity.
	CurrencyCodeCOP = CurrencyCode_CURRENCY_CODE_COP
	//CurrencyCodeCRC is a simple holder for CurrencyCode_CURRENCY_CODE_CRC for simplicity.
	CurrencyCodeCRC = CurrencyCode_CURRENCY_CODE_CRC
	//CurrencyCodeCUP is a simple holder for CurrencyCode_CURRENCY_CODE_CUP for simplicity.
	CurrencyCodeCUP = CurrencyCode_CURRENCY_CODE_CUP
	//CurrencyCodeCVE is a simple holder for CurrencyCode_CURRENCY_CODE_CVE for simplicity.
	CurrencyCodeCVE = CurrencyCode_CURRENCY_CODE_CVE
	//CurrencyCodeCZK is a simple holder for CurrencyCode_CURRENCY_CODE_CZK for simplicity.
	CurrencyCodeCZK = CurrencyCode_CURRENCY_CODE_CZK
	//CurrencyCodeDJF is a simple holder for CurrencyCode_CURRENCY_CODE_DJF for simplicity.
	CurrencyCodeDJF = CurrencyCode_CURRENCY_CODE_DJF
	//CurrencyCodeDKK is a simple holder for CurrencyCode_CURRENCY_CODE_DKK for simplicity.
	CurrencyCodeDKK = CurrencyCode_CURRENCY_CODE_DKK
	//CurrencyCodeDOP is a simple holder for CurrencyCode_CURRENCY_CODE_DOP for simplicity.
	CurrencyCodeDOP = CurrencyCode_CURRENCY_CODE_DOP
	//CurrencyCodeDZD is a simple holder for CurrencyCode_CURRENCY_CODE_DZD for simplicity.
	CurrencyCodeDZD = CurrencyCode_CURRENCY_CODE_DZD
	//CurrencyCodeEGP is a simple holder for CurrencyCode_CURRENCY_CODE_EGP for simplicity.
	CurrencyCodeEGP = CurrencyCode_CURRENCY_CODE_EGP
	//CurrencyCodeERN is a simple holder for CurrencyCode_CURRENCY_CODE_ERN for simplicity.
	CurrencyCodeERN = CurrencyCode_CURRENCY_CODE_ERN
	//CurrencyCodeETB is a simple holder for CurrencyCode_CURRENCY_CODE_ETB for simplicity.
	CurrencyCodeETB = CurrencyCode_CURRENCY_CODE_ETB
	//CurrencyCodeEUR is a simple holder for CurrencyCode_CURRENCY_CODE_EUR for simplicity.
	CurrencyCodeEUR = CurrencyCode_CURRENCY_CODE_EUR
	//CurrencyCodeFJD is a simple holder for CurrencyCode_CURRENCY_CODE_FJD for simplicity.
	CurrencyCodeFJD = CurrencyCode_CURRENCY_CODE_FJD
	//CurrencyCodeFKP is a simple holder for CurrencyCode_CURRENCY_CODE_FKP for simplicity.
	CurrencyCodeFKP = CurrencyCode_CURRENCY_CODE_FKP
	//CurrencyCodeGBP is a simple holder for CurrencyCode_CURRENCY_CODE_GBP for simplicity.
	CurrencyCodeGBP = CurrencyCode_CURRENCY_CODE_GBP
	//CurrencyCodeGEL is a simple holder for CurrencyCode_CURRENCY_CODE_GEL for simplicity.
	CurrencyCodeGEL = CurrencyCode_CURRENCY_CODE_GEL
	//CurrencyCodeGHS is a simple holder for CurrencyCode_CURRENCY_CODE_GHS for simplicity.
	CurrencyCodeGHS = CurrencyCode_CURRENCY_CODE_GHS
	//CurrencyCodeGIP is a simple holder for CurrencyCode_CURRENCY_CODE_GIP for simplicity.
	CurrencyCodeGIP = CurrencyCode_CURRENCY_CODE_GIP
	//CurrencyCodeGMD is a simple holder for CurrencyCode_CURRENCY_CODE_GMD for simplicity.
	CurrencyCodeGMD = CurrencyCode_CURRENCY_CODE_GMD
	//CurrencyCodeGNF is a simple holder for CurrencyCode_CURRENCY_CODE_GNF for simplicity.
	CurrencyCodeGNF = CurrencyCode_CURRENCY_CODE_GNF
	//CurrencyCodeGTQ is a simple holder for CurrencyCode_CURRENCY_CODE_GTQ for simplicity.
	CurrencyCodeGTQ = CurrencyCode_CURRENCY_CODE_GTQ
	//CurrencyCodeGYD is a simple holder for CurrencyCode_CURRENCY_CODE_GYD for simplicity.
	CurrencyCodeGYD = CurrencyCode_CURRENCY_CODE_GYD
	//CurrencyCodeHKD is a simple holder for CurrencyCode_CURRENCY_CODE_HKD for simplicity.
	CurrencyCodeHKD = CurrencyCode_CURRENCY_CODE_HKD
	//CurrencyCodeHNL is a simple holder for CurrencyCode_CURRENCY_CODE_HNL for simplicity.
	CurrencyCodeHNL = CurrencyCode_CURRENCY_CODE_HNL
	//CurrencyCodeHRK is a simple holder for CurrencyCode_CURRENCY_CODE_HRK for simplicity.
	CurrencyCodeHRK = CurrencyCode_CURRENCY_CODE_HRK
	//CurrencyCodeHUF is a simple holder for CurrencyCode_CURRENCY_CODE_HUF for simplicity.
	CurrencyCodeHUF = CurrencyCode_CURRENCY_CODE_HUF
	//CurrencyCodeIDR is a simple holder for CurrencyCode_CURRENCY_CODE_IDR for simplicity.
	CurrencyCodeIDR = CurrencyCode_CURRENCY_CODE_IDR
	//CurrencyCodeILS is a simple holder for CurrencyCode_CURRENCY_CODE_ILS for simplicity.
	CurrencyCodeILS = CurrencyCode_CURRENCY_CODE_ILS
	//CurrencyCodeINR is a simple holder for CurrencyCode_CURRENCY_CODE_INR for simplicity.
	CurrencyCodeINR = CurrencyCode_CURRENCY_CODE_INR
	//CurrencyCodeIQD is a simple holder for CurrencyCode_CURRENCY_CODE_IQD for simplicity.
	CurrencyCodeIQD = CurrencyCode_CURRENCY_CODE_IQD
	//CurrencyCodeIRR is a simple holder for CurrencyCode_CURRENCY_CODE_IRR for simplicity.
	CurrencyCodeIRR = CurrencyCode_CURRENCY_CODE_IRR
	//CurrencyCodeISK is a simple holder for CurrencyCode_CURRENCY_CODE_ISK for simplicity.
	CurrencyCodeISK = CurrencyCode_CURRENCY_CODE_ISK
	//CurrencyCodeJMD is a simple holder for CurrencyCode_CURRENCY_CODE_JMD for simplicity.
	CurrencyCodeJMD = CurrencyCode_CURRENCY_CODE_JMD
	//CurrencyCodeJOD is a simple holder for CurrencyCode_CURRENCY_CODE_JOD for simplicity.
	CurrencyCodeJOD = CurrencyCode_CURRENCY_CODE_JOD
	//CurrencyCodeJPY is a simple holder for CurrencyCode_CURRENCY_CODE_JPY for simplicity.
	CurrencyCodeJPY = CurrencyCode_CURRENCY_CODE_JPY
	//CurrencyCodeKES is a simple holder for CurrencyCode_CURRENCY_CODE_KES for simplicity.
	CurrencyCodeKES = CurrencyCode_CURRENCY_CODE_KES
	//CurrencyCodeKGS is a simple holder for CurrencyCode_CURRENCY_CODE_KGS for simplicity.
	CurrencyCodeKGS = CurrencyCode_CURRENCY_CODE_KGS
	//CurrencyCodeKHR is a simple holder for CurrencyCode_CURRENCY_CODE_KHR for simplicity.
	CurrencyCodeKHR = CurrencyCode_CURRENCY_CODE_KHR
	//CurrencyCodeKMF is a simple holder for CurrencyCode_CURRENCY_CODE_KMF for simplicity.
	CurrencyCodeKMF = CurrencyCode_CURRENCY_CODE_KMF
	//CurrencyCodeKPW is a simple holder for CurrencyCode_CURRENCY_CODE_KPW for simplicity.
	CurrencyCodeKPW = CurrencyCode_CURRENCY_CODE_KPW
	//CurrencyCodeKRW is a simple holder for CurrencyCode_CURRENCY_CODE_KRW for simplicity.
	CurrencyCodeKRW = CurrencyCode_CURRENCY_CODE_KRW
	//CurrencyCodeKWD is a simple holder for CurrencyCode_CURRENCY_CODE_KWD for simplicity.
	CurrencyCodeKWD = CurrencyCode_CURRENCY_CODE_KWD
	//CurrencyCodeKYD is a simple holder for CurrencyCode_CURRENCY_CODE_KYD for simplicity.
	CurrencyCodeKYD = CurrencyCode_CURRENCY_CODE_KYD
	//CurrencyCodeKZT is a simple holder for CurrencyCode_CURRENCY_CODE_KZT for simplicity.
	CurrencyCodeKZT = CurrencyCode_CURRENCY_CODE_KZT
	//CurrencyCodeLAK is a simple holder for CurrencyCode_CURRENCY_CODE_LAK for simplicity.
	CurrencyCodeLAK = CurrencyCode_CURRENCY_CODE_LAK
	//CurrencyCodeLBP is a simple holder for CurrencyCode_CURRENCY_CODE_LBP for simplicity.
	CurrencyCodeLBP = CurrencyCode_CURRENCY_CODE_LBP
	//CurrencyCodeLKR is a simple holder for CurrencyCode_CURRENCY_CODE_LKR for simplicity.
	CurrencyCodeLKR = CurrencyCode_CURRENCY_CODE_LKR
	//CurrencyCodeLRD is a simple holder for CurrencyCode_CURRENCY_CODE_LRD for simplicity.
	CurrencyCodeLRD = CurrencyCode_CURRENCY_CODE_LRD
	//CurrencyCodeLYD is a simple holder for CurrencyCode_CURRENCY_CODE_LYD for simplicity.
	CurrencyCodeLYD = CurrencyCode_CURRENCY_CODE_LYD
	//CurrencyCodeMAD is a simple holder for CurrencyCode_CURRENCY_CODE_MAD for simplicity.
	CurrencyCodeMAD = CurrencyCode_CURRENCY_CODE_MAD
	//CurrencyCodeMDL is a simple holder for CurrencyCode_CURRENCY_CODE_MDL for simplicity.
	CurrencyCodeMDL = CurrencyCode_CURRENCY_CODE_MDL
	//CurrencyCodeMGA is a simple holder for CurrencyCode_CURRENCY_CODE_MGA for simplicity.
	CurrencyCodeMGA = CurrencyCode_CURRENCY_CODE_MGA
	//CurrencyCodeMKD is a simple holder for CurrencyCode_CURRENCY_CODE_MKD for simplicity.
	CurrencyCodeMKD = CurrencyCode_CURRENCY_CODE_MKD
	//CurrencyCodeMMK is a simple holder for CurrencyCode_CURRENCY_CODE_MMK for simplicity.
	CurrencyCodeMMK = CurrencyCode_CURRENCY_CODE_MMK
	//CurrencyCodeMNT is a simple holder for CurrencyCode_CURRENCY_CODE_MNT for simplicity.
	CurrencyCodeMNT = CurrencyCode_CURRENCY_CODE_MNT
	//CurrencyCodeMOP is a simple holder for CurrencyCode_CURRENCY_CODE_MOP for simplicity.
	CurrencyCodeMOP = CurrencyCode_CURRENCY_CODE_MOP
	//CurrencyCodeMRO is a simple holder for CurrencyCode_CURRENCY_CODE_MRO for simplicity.
	CurrencyCodeMRO = CurrencyCode_CURRENCY_CODE_MRO
	//CurrencyCodeMUR is a simple holder for CurrencyCode_CURRENCY_CODE_MUR for simplicity.
	CurrencyCodeMUR = CurrencyCode_CURRENCY_CODE_MUR
	//CurrencyCodeMVR is a simple holder for CurrencyCode_CURRENCY_CODE_MVR for simplicity.
	CurrencyCodeMVR = CurrencyCode_CURRENCY_CODE_MVR
	//CurrencyCodeMWK is a simple holder for CurrencyCode_CURRENCY_CODE_MWK for simplicity.
	CurrencyCodeMWK = CurrencyCode_CURRENCY_CODE_MWK
	//CurrencyCodeMXN is a simple holder for CurrencyCode_CURRENCY_CODE_MXN for simplicity.
	CurrencyCodeMXN = CurrencyCode_CURRENCY_CODE_MXN
	//CurrencyCodeMYR is a simple holder for CurrencyCode_CURRENCY_CODE_MYR for simplicity.
	CurrencyCodeMYR = CurrencyCode_CURRENCY_CODE_MYR
	//CurrencyCodeMZN is a simple holder for CurrencyCode_CURRENCY_CODE_MZN for simplicity.
	CurrencyCodeMZN = CurrencyCode_CURRENCY_CODE_MZN
	//CurrencyCodeNGN is a simple holder for CurrencyCode_CURRENCY_CODE_NGN for simplicity.
	CurrencyCodeNGN = CurrencyCode_CURRENCY_CODE_NGN
	//CurrencyCodeNIO is a simple holder for CurrencyCode_CURRENCY_CODE_NIO for simplicity.
	CurrencyCodeNIO = CurrencyCode_CURRENCY_CODE_NIO
	//CurrencyCodeNOK is a simple holder for CurrencyCode_CURRENCY_CODE_NOK for simplicity.
	CurrencyCodeNOK = CurrencyCode_CURRENCY_CODE_NOK
	//CurrencyCodeNPR is a simple holder for CurrencyCode_CURRENCY_CODE_NPR for simplicity.
	CurrencyCodeNPR = CurrencyCode_CURRENCY_CODE_NPR
	//CurrencyCodeNZD is a simple holder for CurrencyCode_CURRENCY_CODE_NZD for simplicity.
	CurrencyCodeNZD = CurrencyCode_CURRENCY_CODE_NZD
	//CurrencyCodeOMR is a simple holder for CurrencyCode_CURRENCY_CODE_OMR for simplicity.
	CurrencyCodeOMR = CurrencyCode_CURRENCY_CODE_OMR
	//CurrencyCodePEN is a simple holder for CurrencyCode_CURRENCY_CODE_PEN for simplicity.
	CurrencyCodePEN = CurrencyCode_CURRENCY_CODE_PEN
	//CurrencyCodePGK is a simple holder for CurrencyCode_CURRENCY_CODE_PGK for simplicity.
	CurrencyCodePGK = CurrencyCode_CURRENCY_CODE_PGK
	//CurrencyCodePHP is a simple holder for CurrencyCode_CURRENCY_CODE_PHP for simplicity.
	CurrencyCodePHP = CurrencyCode_CURRENCY_CODE_PHP
	//CurrencyCodePKR is a simple holder for CurrencyCode_CURRENCY_CODE_PKR for simplicity.
	CurrencyCodePKR = CurrencyCode_CURRENCY_CODE_PKR
	//CurrencyCodePLN is a simple holder for CurrencyCode_CURRENCY_CODE_PLN for simplicity.
	CurrencyCodePLN = CurrencyCode_CURRENCY_CODE_PLN
	//CurrencyCodePYG is a simple holder for CurrencyCode_CURRENCY_CODE_PYG for simplicity.
	CurrencyCodePYG = CurrencyCode_CURRENCY_CODE_PYG
	//CurrencyCodeQAR is a simple holder for CurrencyCode_CURRENCY_CODE_QAR for simplicity.
	CurrencyCodeQAR = CurrencyCode_CURRENCY_CODE_QAR
	//CurrencyCodeRON is a simple holder for CurrencyCode_CURRENCY_CODE_RON for simplicity.
	CurrencyCodeRON = CurrencyCode_CURRENCY_CODE_RON
	//CurrencyCodeRSD is a simple holder for CurrencyCode_CURRENCY_CODE_RSD for simplicity.
	CurrencyCodeRSD = CurrencyCode_CURRENCY_CODE_RSD
	//CurrencyCodeRUB is a simple holder for CurrencyCode_CURRENCY_CODE_RUB for simplicity.
	CurrencyCodeRUB = CurrencyCode_CURRENCY_CODE_RUB
	//CurrencyCodeRWF is a simple holder for CurrencyCode_CURRENCY_CODE_RWF for simplicity.
	CurrencyCodeRWF = CurrencyCode_CURRENCY_CODE_RWF
	//CurrencyCodeSAR is a simple holder for CurrencyCode_CURRENCY_CODE_SAR for simplicity.
	CurrencyCodeSAR = CurrencyCode_CURRENCY_CODE_SAR
	//CurrencyCodeSBD is a simple holder for CurrencyCode_CURRENCY_CODE_SBD for simplicity.
	CurrencyCodeSBD = CurrencyCode_CURRENCY_CODE_SBD
	//CurrencyCodeSCR is a simple holder for CurrencyCode_CURRENCY_CODE_SCR for simplicity.
	CurrencyCodeSCR = CurrencyCode_CURRENCY_CODE_SCR
	//CurrencyCodeSDG is a simple holder for CurrencyCode_CURRENCY_CODE_SDG for simplicity.
	CurrencyCodeSDG = CurrencyCode_CURRENCY_CODE_SDG
	//CurrencyCodeSEK is a simple holder for CurrencyCode_CURRENCY_CODE_SEK for simplicity.
	CurrencyCodeSEK = CurrencyCode_CURRENCY_CODE_SEK
	//CurrencyCodeSGD is a simple holder for CurrencyCode_CURRENCY_CODE_SGD for simplicity.
	CurrencyCodeSGD = CurrencyCode_CURRENCY_CODE_SGD
	//CurrencyCodeSHP is a simple holder for CurrencyCode_CURRENCY_CODE_SHP for simplicity.
	CurrencyCodeSHP = CurrencyCode_CURRENCY_CODE_SHP
	//CurrencyCodeSLL is a simple holder for CurrencyCode_CURRENCY_CODE_SLL for simplicity.
	CurrencyCodeSLL = CurrencyCode_CURRENCY_CODE_SLL
	//CurrencyCodeSOS is a simple holder for CurrencyCode_CURRENCY_CODE_SOS for simplicity.
	CurrencyCodeSOS = CurrencyCode_CURRENCY_CODE_SOS
	//CurrencyCodeSRD is a simple holder for CurrencyCode_CURRENCY_CODE_SRD for simplicity.
	CurrencyCodeSRD = CurrencyCode_CURRENCY_CODE_SRD
	//CurrencyCodeSSP is a simple holder for CurrencyCode_CURRENCY_CODE_SSP for simplicity.
	CurrencyCodeSSP = CurrencyCode_CURRENCY_CODE_SSP
	//CurrencyCodeSTD is a simple holder for CurrencyCode_CURRENCY_CODE_STD for simplicity.
	CurrencyCodeSTD = CurrencyCode_CURRENCY_CODE_STD
	//CurrencyCodeSYP is a simple holder for CurrencyCode_CURRENCY_CODE_SYP for simplicity.
	CurrencyCodeSYP = CurrencyCode_CURRENCY_CODE_SYP
	//CurrencyCodeSZL is a simple holder for CurrencyCode_CURRENCY_CODE_SZL for simplicity.
	CurrencyCodeSZL = CurrencyCode_CURRENCY_CODE_SZL
	//CurrencyCodeTHB is a simple holder for CurrencyCode_CURRENCY_CODE_THB for simplicity.
	CurrencyCodeTHB = CurrencyCode_CURRENCY_CODE_THB
	//CurrencyCodeTJS is a simple holder for CurrencyCode_CURRENCY_CODE_TJS for simplicity.
	CurrencyCodeTJS = CurrencyCode_CURRENCY_CODE_TJS
	//CurrencyCodeTMT is a simple holder for CurrencyCode_CURRENCY_CODE_TMT for simplicity.
	CurrencyCodeTMT = CurrencyCode_CURRENCY_CODE_TMT
	//CurrencyCodeTND is a simple holder for CurrencyCode_CURRENCY_CODE_TND for simplicity.
	CurrencyCodeTND = CurrencyCode_CURRENCY_CODE_TND
	//CurrencyCodeTOP is a simple holder for CurrencyCode_CURRENCY_CODE_TOP for simplicity.
	CurrencyCodeTOP = CurrencyCode_CURRENCY_CODE_TOP
	//CurrencyCodeTRY is a simple holder for CurrencyCode_CURRENCY_CODE_TRY for simplicity.
	CurrencyCodeTRY = CurrencyCode_CURRENCY_CODE_TRY
	//CurrencyCodeTTD is a simple holder for CurrencyCode_CURRENCY_CODE_TTD for simplicity.
	CurrencyCodeTTD = CurrencyCode_CURRENCY_CODE_TTD
	//CurrencyCodeTWD is a simple holder for CurrencyCode_CURRENCY_CODE_TWD for simplicity.
	CurrencyCodeTWD = CurrencyCode_CURRENCY_CODE_TWD
	//CurrencyCodeTZS is a simple holder for CurrencyCode_CURRENCY_CODE_TZS for simplicity.
	CurrencyCodeTZS = CurrencyCode_CURRENCY_CODE_TZS
	//CurrencyCodeUAH is a simple holder for CurrencyCode_CURRENCY_CODE_UAH for simplicity.
	CurrencyCodeUAH = CurrencyCode_CURRENCY_CODE_UAH
	//CurrencyCodeUGX is a simple holder for CurrencyCode_CURRENCY_CODE_UGX for simplicity.
	CurrencyCodeUGX = CurrencyCode_CURRENCY_CODE_UGX
	//CurrencyCodeUSD is a simple holder for CurrencyCode_CURRENCY_CODE_USD for simplicity.
	CurrencyCodeUSD = CurrencyCode_CURRENCY_CODE_USD
	//CurrencyCodeUYU is a simple holder for CurrencyCode_CURRENCY_CODE_UYU for simplicity.
	CurrencyCodeUYU = CurrencyCode_CURRENCY_CODE_UYU
	//CurrencyCodeUZS is a simple holder for CurrencyCode_CURRENCY_CODE_UZS for simplicity.
	CurrencyCodeUZS = CurrencyCode_CURRENCY_CODE_UZS
	//CurrencyCodeVEF is a simple holder for CurrencyCode_CURRENCY_CODE_VEF for simplicity.
	CurrencyCodeVEF = CurrencyCode_CURRENCY_CODE_VEF
	//CurrencyCodeVND is a simple holder for CurrencyCode_CURRENCY_CODE_VND for simplicity.
	CurrencyCodeVND = CurrencyCode_CURRENCY_CODE_VND
	//CurrencyCodeVUV is a simple holder for CurrencyCode_CURRENCY_CODE_VUV for simplicity.
	CurrencyCodeVUV = CurrencyCode_CURRENCY_CODE_VUV
	//CurrencyCodeWST is a simple holder for CurrencyCode_CURRENCY_CODE_WST for simplicity.
	CurrencyCodeWST = CurrencyCode_CURRENCY_CODE_WST
	//CurrencyCodeXAF is a simple holder for CurrencyCode_CURRENCY_CODE_XAF for simplicity.
	CurrencyCodeXAF = CurrencyCode_CURRENCY_CODE_XAF
	//CurrencyCodeXCD is a simple holder for CurrencyCode_CURRENCY_CODE_XCD for simplicity.
	CurrencyCodeXCD = CurrencyCode_CURRENCY_CODE_XCD
	//CurrencyCodeXOF is a simple holder for CurrencyCode_CURRENCY_CODE_XOF for simplicity.
	CurrencyCodeXOF = CurrencyCode_CURRENCY_CODE_XOF
	//CurrencyCodeXPF is a simple holder for CurrencyCode_CURRENCY_CODE_XPF for simplicity.
	CurrencyCodeXPF = CurrencyCode_CURRENCY_CODE_XPF
	//CurrencyCodeYER is a simple holder for CurrencyCode_CURRENCY_CODE_YER for simplicity.
	CurrencyCodeYER = CurrencyCode_CURRENCY_CODE_YER
	//CurrencyCodeZAR is a simple holder for CurrencyCode_CURRENCY_CODE_ZAR for simplicity.
	CurrencyCodeZAR = CurrencyCode_CURRENCY_CODE_ZAR
	//CurrencyCodeZMW is a simple holder for CurrencyCode_CURRENCY_CODE_ZMW for simplicity.
	CurrencyCodeZMW = CurrencyCode_CURRENCY_CODE_ZMW
	//CurrencyCodeZWL is a simple holder for CurrencyCode_CURRENCY_CODE_ZWL for simplicity.
	CurrencyCodeZWL = CurrencyCode_CURRENCY_CODE_ZWL

	// CurrencyCodeToCurrency is a map from CurrencyCode to Currency.
	CurrencyCodeToCurrency = map[CurrencyCode]*Currency{
		CurrencyCodeAED: currency1,
		CurrencyCodeAFN: currency2,
		CurrencyCodeALL: currency3,
		CurrencyCodeAMD: currency4,
		CurrencyCodeANG: currency5,
		CurrencyCodeAOA: currency6,
		CurrencyCodeARS: currency7,
		CurrencyCodeAUD: currency8,
		CurrencyCodeAWG: currency9,
		CurrencyCodeAZN: currency10,
		CurrencyCodeBAM: currency11,
		CurrencyCodeBBD: currency12,
		CurrencyCodeBDT: currency13,
		CurrencyCodeBGN: currency14,
		CurrencyCodeBHD: currency15,
		CurrencyCodeBIF: currency16,
		CurrencyCodeBMD: currency17,
		CurrencyCodeBND: currency18,
		CurrencyCodeBOB: currency19,
		CurrencyCodeBRL: currency20,
		CurrencyCodeBSD: currency21,
		CurrencyCodeBWP: currency22,
		CurrencyCodeBYR: currency23,
		CurrencyCodeBZD: currency24,
		CurrencyCodeCAD: currency25,
		CurrencyCodeCHF: currency26,
		CurrencyCodeCLP: currency27,
		CurrencyCodeCNY: currency28,
		CurrencyCodeCOP: currency29,
		CurrencyCodeCRC: currency30,
		CurrencyCodeCUP: currency31,
		CurrencyCodeCVE: currency32,
		CurrencyCodeCZK: currency33,
		CurrencyCodeDJF: currency34,
		CurrencyCodeDKK: currency35,
		CurrencyCodeDOP: currency36,
		CurrencyCodeDZD: currency37,
		CurrencyCodeEGP: currency38,
		CurrencyCodeERN: currency39,
		CurrencyCodeETB: currency40,
		CurrencyCodeEUR: currency41,
		CurrencyCodeFJD: currency42,
		CurrencyCodeFKP: currency43,
		CurrencyCodeGBP: currency44,
		CurrencyCodeGEL: currency45,
		CurrencyCodeGHS: currency46,
		CurrencyCodeGIP: currency47,
		CurrencyCodeGMD: currency48,
		CurrencyCodeGNF: currency49,
		CurrencyCodeGTQ: currency50,
		CurrencyCodeGYD: currency51,
		CurrencyCodeHKD: currency52,
		CurrencyCodeHNL: currency53,
		CurrencyCodeHRK: currency54,
		CurrencyCodeHUF: currency55,
		CurrencyCodeIDR: currency56,
		CurrencyCodeILS: currency57,
		CurrencyCodeINR: currency58,
		CurrencyCodeIQD: currency59,
		CurrencyCodeIRR: currency60,
		CurrencyCodeISK: currency61,
		CurrencyCodeJMD: currency62,
		CurrencyCodeJOD: currency63,
		CurrencyCodeJPY: currency64,
		CurrencyCodeKES: currency65,
		CurrencyCodeKGS: currency66,
		CurrencyCodeKHR: currency67,
		CurrencyCodeKMF: currency68,
		CurrencyCodeKPW: currency69,
		CurrencyCodeKRW: currency70,
		CurrencyCodeKWD: currency71,
		CurrencyCodeKYD: currency72,
		CurrencyCodeKZT: currency73,
		CurrencyCodeLAK: currency74,
		CurrencyCodeLBP: currency75,
		CurrencyCodeLKR: currency76,
		CurrencyCodeLRD: currency77,
		CurrencyCodeLYD: currency78,
		CurrencyCodeMAD: currency79,
		CurrencyCodeMDL: currency80,
		CurrencyCodeMGA: currency81,
		CurrencyCodeMKD: currency82,
		CurrencyCodeMMK: currency83,
		CurrencyCodeMNT: currency84,
		CurrencyCodeMOP: currency85,
		CurrencyCodeMRO: currency86,
		CurrencyCodeMUR: currency87,
		CurrencyCodeMVR: currency88,
		CurrencyCodeMWK: currency89,
		CurrencyCodeMXN: currency90,
		CurrencyCodeMYR: currency91,
		CurrencyCodeMZN: currency92,
		CurrencyCodeNGN: currency93,
		CurrencyCodeNIO: currency94,
		CurrencyCodeNOK: currency95,
		CurrencyCodeNPR: currency96,
		CurrencyCodeNZD: currency97,
		CurrencyCodeOMR: currency98,
		CurrencyCodePEN: currency99,
		CurrencyCodePGK: currency100,
		CurrencyCodePHP: currency101,
		CurrencyCodePKR: currency102,
		CurrencyCodePLN: currency103,
		CurrencyCodePYG: currency104,
		CurrencyCodeQAR: currency105,
		CurrencyCodeRON: currency106,
		CurrencyCodeRSD: currency107,
		CurrencyCodeRUB: currency108,
		CurrencyCodeRWF: currency109,
		CurrencyCodeSAR: currency110,
		CurrencyCodeSBD: currency111,
		CurrencyCodeSCR: currency112,
		CurrencyCodeSDG: currency113,
		CurrencyCodeSEK: currency114,
		CurrencyCodeSGD: currency115,
		CurrencyCodeSHP: currency116,
		CurrencyCodeSLL: currency117,
		CurrencyCodeSOS: currency118,
		CurrencyCodeSRD: currency119,
		CurrencyCodeSSP: currency120,
		CurrencyCodeSTD: currency121,
		CurrencyCodeSYP: currency122,
		CurrencyCodeSZL: currency123,
		CurrencyCodeTHB: currency124,
		CurrencyCodeTJS: currency125,
		CurrencyCodeTMT: currency126,
		CurrencyCodeTND: currency127,
		CurrencyCodeTOP: currency128,
		CurrencyCodeTRY: currency129,
		CurrencyCodeTTD: currency130,
		CurrencyCodeTWD: currency131,
		CurrencyCodeTZS: currency132,
		CurrencyCodeUAH: currency133,
		CurrencyCodeUGX: currency134,
		CurrencyCodeUSD: currency135,
		CurrencyCodeUYU: currency136,
		CurrencyCodeUZS: currency137,
		CurrencyCodeVEF: currency138,
		CurrencyCodeVND: currency139,
		CurrencyCodeVUV: currency140,
		CurrencyCodeWST: currency141,
		CurrencyCodeXAF: currency142,
		CurrencyCodeXCD: currency143,
		CurrencyCodeXOF: currency144,
		CurrencyCodeXPF: currency145,
		CurrencyCodeYER: currency146,
		CurrencyCodeZAR: currency147,
		CurrencyCodeZMW: currency148,
		CurrencyCodeZWL: currency149,
	}

	// CurrencyCodeToSimpleString is a map from CurrencyCode to simple string.
	CurrencyCodeToSimpleString = map[CurrencyCode]string{
		CurrencyCodeAED: "AED",
		CurrencyCodeAFN: "AFN",
		CurrencyCodeALL: "ALL",
		CurrencyCodeAMD: "AMD",
		CurrencyCodeANG: "ANG",
		CurrencyCodeAOA: "AOA",
		CurrencyCodeARS: "ARS",
		CurrencyCodeAUD: "AUD",
		CurrencyCodeAWG: "AWG",
		CurrencyCodeAZN: "AZN",
		CurrencyCodeBAM: "BAM",
		CurrencyCodeBBD: "BBD",
		CurrencyCodeBDT: "BDT",
		CurrencyCodeBGN: "BGN",
		CurrencyCodeBHD: "BHD",
		CurrencyCodeBIF: "BIF",
		CurrencyCodeBMD: "BMD",
		CurrencyCodeBND: "BND",
		CurrencyCodeBOB: "BOB",
		CurrencyCodeBRL: "BRL",
		CurrencyCodeBSD: "BSD",
		CurrencyCodeBWP: "BWP",
		CurrencyCodeBYR: "BYR",
		CurrencyCodeBZD: "BZD",
		CurrencyCodeCAD: "CAD",
		CurrencyCodeCHF: "CHF",
		CurrencyCodeCLP: "CLP",
		CurrencyCodeCNY: "CNY",
		CurrencyCodeCOP: "COP",
		CurrencyCodeCRC: "CRC",
		CurrencyCodeCUP: "CUP",
		CurrencyCodeCVE: "CVE",
		CurrencyCodeCZK: "CZK",
		CurrencyCodeDJF: "DJF",
		CurrencyCodeDKK: "DKK",
		CurrencyCodeDOP: "DOP",
		CurrencyCodeDZD: "DZD",
		CurrencyCodeEGP: "EGP",
		CurrencyCodeERN: "ERN",
		CurrencyCodeETB: "ETB",
		CurrencyCodeEUR: "EUR",
		CurrencyCodeFJD: "FJD",
		CurrencyCodeFKP: "FKP",
		CurrencyCodeGBP: "GBP",
		CurrencyCodeGEL: "GEL",
		CurrencyCodeGHS: "GHS",
		CurrencyCodeGIP: "GIP",
		CurrencyCodeGMD: "GMD",
		CurrencyCodeGNF: "GNF",
		CurrencyCodeGTQ: "GTQ",
		CurrencyCodeGYD: "GYD",
		CurrencyCodeHKD: "HKD",
		CurrencyCodeHNL: "HNL",
		CurrencyCodeHRK: "HRK",
		CurrencyCodeHUF: "HUF",
		CurrencyCodeIDR: "IDR",
		CurrencyCodeILS: "ILS",
		CurrencyCodeINR: "INR",
		CurrencyCodeIQD: "IQD",
		CurrencyCodeIRR: "IRR",
		CurrencyCodeISK: "ISK",
		CurrencyCodeJMD: "JMD",
		CurrencyCodeJOD: "JOD",
		CurrencyCodeJPY: "JPY",
		CurrencyCodeKES: "KES",
		CurrencyCodeKGS: "KGS",
		CurrencyCodeKHR: "KHR",
		CurrencyCodeKMF: "KMF",
		CurrencyCodeKPW: "KPW",
		CurrencyCodeKRW: "KRW",
		CurrencyCodeKWD: "KWD",
		CurrencyCodeKYD: "KYD",
		CurrencyCodeKZT: "KZT",
		CurrencyCodeLAK: "LAK",
		CurrencyCodeLBP: "LBP",
		CurrencyCodeLKR: "LKR",
		CurrencyCodeLRD: "LRD",
		CurrencyCodeLYD: "LYD",
		CurrencyCodeMAD: "MAD",
		CurrencyCodeMDL: "MDL",
		CurrencyCodeMGA: "MGA",
		CurrencyCodeMKD: "MKD",
		CurrencyCodeMMK: "MMK",
		CurrencyCodeMNT: "MNT",
		CurrencyCodeMOP: "MOP",
		CurrencyCodeMRO: "MRO",
		CurrencyCodeMUR: "MUR",
		CurrencyCodeMVR: "MVR",
		CurrencyCodeMWK: "MWK",
		CurrencyCodeMXN: "MXN",
		CurrencyCodeMYR: "MYR",
		CurrencyCodeMZN: "MZN",
		CurrencyCodeNGN: "NGN",
		CurrencyCodeNIO: "NIO",
		CurrencyCodeNOK: "NOK",
		CurrencyCodeNPR: "NPR",
		CurrencyCodeNZD: "NZD",
		CurrencyCodeOMR: "OMR",
		CurrencyCodePEN: "PEN",
		CurrencyCodePGK: "PGK",
		CurrencyCodePHP: "PHP",
		CurrencyCodePKR: "PKR",
		CurrencyCodePLN: "PLN",
		CurrencyCodePYG: "PYG",
		CurrencyCodeQAR: "QAR",
		CurrencyCodeRON: "RON",
		CurrencyCodeRSD: "RSD",
		CurrencyCodeRUB: "RUB",
		CurrencyCodeRWF: "RWF",
		CurrencyCodeSAR: "SAR",
		CurrencyCodeSBD: "SBD",
		CurrencyCodeSCR: "SCR",
		CurrencyCodeSDG: "SDG",
		CurrencyCodeSEK: "SEK",
		CurrencyCodeSGD: "SGD",
		CurrencyCodeSHP: "SHP",
		CurrencyCodeSLL: "SLL",
		CurrencyCodeSOS: "SOS",
		CurrencyCodeSRD: "SRD",
		CurrencyCodeSSP: "SSP",
		CurrencyCodeSTD: "STD",
		CurrencyCodeSYP: "SYP",
		CurrencyCodeSZL: "SZL",
		CurrencyCodeTHB: "THB",
		CurrencyCodeTJS: "TJS",
		CurrencyCodeTMT: "TMT",
		CurrencyCodeTND: "TND",
		CurrencyCodeTOP: "TOP",
		CurrencyCodeTRY: "TRY",
		CurrencyCodeTTD: "TTD",
		CurrencyCodeTWD: "TWD",
		CurrencyCodeTZS: "TZS",
		CurrencyCodeUAH: "UAH",
		CurrencyCodeUGX: "UGX",
		CurrencyCodeUSD: "USD",
		CurrencyCodeUYU: "UYU",
		CurrencyCodeUZS: "UZS",
		CurrencyCodeVEF: "VEF",
		CurrencyCodeVND: "VND",
		CurrencyCodeVUV: "VUV",
		CurrencyCodeWST: "WST",
		CurrencyCodeXAF: "XAF",
		CurrencyCodeXCD: "XCD",
		CurrencyCodeXOF: "XOF",
		CurrencyCodeXPF: "XPF",
		CurrencyCodeYER: "YER",
		CurrencyCodeZAR: "ZAR",
		CurrencyCodeZMW: "ZMW",
		CurrencyCodeZWL: "ZWL",
	}

	// SimpleStringToCurrencyCode is a map from simple string to CurrencyCode.
	SimpleStringToCurrencyCode = map[string]CurrencyCode{
		"AED": CurrencyCodeAED,
		"AFN": CurrencyCodeAFN,
		"ALL": CurrencyCodeALL,
		"AMD": CurrencyCodeAMD,
		"ANG": CurrencyCodeANG,
		"AOA": CurrencyCodeAOA,
		"ARS": CurrencyCodeARS,
		"AUD": CurrencyCodeAUD,
		"AWG": CurrencyCodeAWG,
		"AZN": CurrencyCodeAZN,
		"BAM": CurrencyCodeBAM,
		"BBD": CurrencyCodeBBD,
		"BDT": CurrencyCodeBDT,
		"BGN": CurrencyCodeBGN,
		"BHD": CurrencyCodeBHD,
		"BIF": CurrencyCodeBIF,
		"BMD": CurrencyCodeBMD,
		"BND": CurrencyCodeBND,
		"BOB": CurrencyCodeBOB,
		"BRL": CurrencyCodeBRL,
		"BSD": CurrencyCodeBSD,
		"BWP": CurrencyCodeBWP,
		"BYR": CurrencyCodeBYR,
		"BZD": CurrencyCodeBZD,
		"CAD": CurrencyCodeCAD,
		"CHF": CurrencyCodeCHF,
		"CLP": CurrencyCodeCLP,
		"CNY": CurrencyCodeCNY,
		"COP": CurrencyCodeCOP,
		"CRC": CurrencyCodeCRC,
		"CUP": CurrencyCodeCUP,
		"CVE": CurrencyCodeCVE,
		"CZK": CurrencyCodeCZK,
		"DJF": CurrencyCodeDJF,
		"DKK": CurrencyCodeDKK,
		"DOP": CurrencyCodeDOP,
		"DZD": CurrencyCodeDZD,
		"EGP": CurrencyCodeEGP,
		"ERN": CurrencyCodeERN,
		"ETB": CurrencyCodeETB,
		"EUR": CurrencyCodeEUR,
		"FJD": CurrencyCodeFJD,
		"FKP": CurrencyCodeFKP,
		"GBP": CurrencyCodeGBP,
		"GEL": CurrencyCodeGEL,
		"GHS": CurrencyCodeGHS,
		"GIP": CurrencyCodeGIP,
		"GMD": CurrencyCodeGMD,
		"GNF": CurrencyCodeGNF,
		"GTQ": CurrencyCodeGTQ,
		"GYD": CurrencyCodeGYD,
		"HKD": CurrencyCodeHKD,
		"HNL": CurrencyCodeHNL,
		"HRK": CurrencyCodeHRK,
		"HUF": CurrencyCodeHUF,
		"IDR": CurrencyCodeIDR,
		"ILS": CurrencyCodeILS,
		"INR": CurrencyCodeINR,
		"IQD": CurrencyCodeIQD,
		"IRR": CurrencyCodeIRR,
		"ISK": CurrencyCodeISK,
		"JMD": CurrencyCodeJMD,
		"JOD": CurrencyCodeJOD,
		"JPY": CurrencyCodeJPY,
		"KES": CurrencyCodeKES,
		"KGS": CurrencyCodeKGS,
		"KHR": CurrencyCodeKHR,
		"KMF": CurrencyCodeKMF,
		"KPW": CurrencyCodeKPW,
		"KRW": CurrencyCodeKRW,
		"KWD": CurrencyCodeKWD,
		"KYD": CurrencyCodeKYD,
		"KZT": CurrencyCodeKZT,
		"LAK": CurrencyCodeLAK,
		"LBP": CurrencyCodeLBP,
		"LKR": CurrencyCodeLKR,
		"LRD": CurrencyCodeLRD,
		"LYD": CurrencyCodeLYD,
		"MAD": CurrencyCodeMAD,
		"MDL": CurrencyCodeMDL,
		"MGA": CurrencyCodeMGA,
		"MKD": CurrencyCodeMKD,
		"MMK": CurrencyCodeMMK,
		"MNT": CurrencyCodeMNT,
		"MOP": CurrencyCodeMOP,
		"MRO": CurrencyCodeMRO,
		"MUR": CurrencyCodeMUR,
		"MVR": CurrencyCodeMVR,
		"MWK": CurrencyCodeMWK,
		"MXN": CurrencyCodeMXN,
		"MYR": CurrencyCodeMYR,
		"MZN": CurrencyCodeMZN,
		"NGN": CurrencyCodeNGN,
		"NIO": CurrencyCodeNIO,
		"NOK": CurrencyCodeNOK,
		"NPR": CurrencyCodeNPR,
		"NZD": CurrencyCodeNZD,
		"OMR": CurrencyCodeOMR,
		"PEN": CurrencyCodePEN,
		"PGK": CurrencyCodePGK,
		"PHP": CurrencyCodePHP,
		"PKR": CurrencyCodePKR,
		"PLN": CurrencyCodePLN,
		"PYG": CurrencyCodePYG,
		"QAR": CurrencyCodeQAR,
		"RON": CurrencyCodeRON,
		"RSD": CurrencyCodeRSD,
		"RUB": CurrencyCodeRUB,
		"RWF": CurrencyCodeRWF,
		"SAR": CurrencyCodeSAR,
		"SBD": CurrencyCodeSBD,
		"SCR": CurrencyCodeSCR,
		"SDG": CurrencyCodeSDG,
		"SEK": CurrencyCodeSEK,
		"SGD": CurrencyCodeSGD,
		"SHP": CurrencyCodeSHP,
		"SLL": CurrencyCodeSLL,
		"SOS": CurrencyCodeSOS,
		"SRD": CurrencyCodeSRD,
		"SSP": CurrencyCodeSSP,
		"STD": CurrencyCodeSTD,
		"SYP": CurrencyCodeSYP,
		"SZL": CurrencyCodeSZL,
		"THB": CurrencyCodeTHB,
		"TJS": CurrencyCodeTJS,
		"TMT": CurrencyCodeTMT,
		"TND": CurrencyCodeTND,
		"TOP": CurrencyCodeTOP,
		"TRY": CurrencyCodeTRY,
		"TTD": CurrencyCodeTTD,
		"TWD": CurrencyCodeTWD,
		"TZS": CurrencyCodeTZS,
		"UAH": CurrencyCodeUAH,
		"UGX": CurrencyCodeUGX,
		"USD": CurrencyCodeUSD,
		"UYU": CurrencyCodeUYU,
		"UZS": CurrencyCodeUZS,
		"VEF": CurrencyCodeVEF,
		"VND": CurrencyCodeVND,
		"VUV": CurrencyCodeVUV,
		"WST": CurrencyCodeWST,
		"XAF": CurrencyCodeXAF,
		"XCD": CurrencyCodeXCD,
		"XOF": CurrencyCodeXOF,
		"XPF": CurrencyCodeXPF,
		"YER": CurrencyCodeYER,
		"ZAR": CurrencyCodeZAR,
		"ZMW": CurrencyCodeZMW,
		"ZWL": CurrencyCodeZWL,
	}

	currency1 = &Currency{
		Name:        "UAE Dirham",
		Code:        CurrencyCodeAED,
		NumericCode: 784,
		MinorUnit:   2,
	}
	currency2 = &Currency{
		Name:        "Afghani",
		Code:        CurrencyCodeAFN,
		NumericCode: 971,
		MinorUnit:   2,
	}
	currency3 = &Currency{
		Name:        "Lek",
		Code:        CurrencyCodeALL,
		NumericCode: 8,
		MinorUnit:   2,
	}
	currency4 = &Currency{
		Name:        "Armenian Dram",
		Code:        CurrencyCodeAMD,
		NumericCode: 51,
		MinorUnit:   2,
	}
	currency5 = &Currency{
		Name:        "Netherlands Antillean Guilder",
		Code:        CurrencyCodeANG,
		NumericCode: 532,
		MinorUnit:   2,
	}
	currency6 = &Currency{
		Name:        "Kwanza",
		Code:        CurrencyCodeAOA,
		NumericCode: 973,
		MinorUnit:   2,
	}
	currency7 = &Currency{
		Name:        "Argentine Peso",
		Code:        CurrencyCodeARS,
		NumericCode: 32,
		MinorUnit:   2,
	}
	currency8 = &Currency{
		Name:        "Australian Dollar",
		Code:        CurrencyCodeAUD,
		NumericCode: 36,
		MinorUnit:   2,
	}
	currency9 = &Currency{
		Name:        "Aruban Florin",
		Code:        CurrencyCodeAWG,
		NumericCode: 533,
		MinorUnit:   2,
	}
	currency10 = &Currency{
		Name:        "Azerbaijanian Manat",
		Code:        CurrencyCodeAZN,
		NumericCode: 944,
		MinorUnit:   2,
	}
	currency11 = &Currency{
		Name:        "Convertible Mark",
		Code:        CurrencyCodeBAM,
		NumericCode: 977,
		MinorUnit:   2,
	}
	currency12 = &Currency{
		Name:        "Barbados Dollar",
		Code:        CurrencyCodeBBD,
		NumericCode: 52,
		MinorUnit:   2,
	}
	currency13 = &Currency{
		Name:        "Taka",
		Code:        CurrencyCodeBDT,
		NumericCode: 50,
		MinorUnit:   2,
	}
	currency14 = &Currency{
		Name:        "Bulgarian Lev",
		Code:        CurrencyCodeBGN,
		NumericCode: 975,
		MinorUnit:   2,
	}
	currency15 = &Currency{
		Name:        "Bahraini Dinar",
		Code:        CurrencyCodeBHD,
		NumericCode: 48,
		MinorUnit:   3,
	}
	currency16 = &Currency{
		Name:        "Burundi Franc",
		Code:        CurrencyCodeBIF,
		NumericCode: 108,
		MinorUnit:   0,
	}
	currency17 = &Currency{
		Name:        "Bermudian Dollar",
		Code:        CurrencyCodeBMD,
		NumericCode: 60,
		MinorUnit:   2,
	}
	currency18 = &Currency{
		Name:        "Brunei Dollar",
		Code:        CurrencyCodeBND,
		NumericCode: 96,
		MinorUnit:   2,
	}
	currency19 = &Currency{
		Name:        "Boliviano",
		Code:        CurrencyCodeBOB,
		NumericCode: 68,
		MinorUnit:   2,
	}
	currency20 = &Currency{
		Name:        "Brazilian Real",
		Code:        CurrencyCodeBRL,
		NumericCode: 986,
		MinorUnit:   2,
	}
	currency21 = &Currency{
		Name:        "Bahamian Dollar",
		Code:        CurrencyCodeBSD,
		NumericCode: 44,
		MinorUnit:   2,
	}
	currency22 = &Currency{
		Name:        "Pula",
		Code:        CurrencyCodeBWP,
		NumericCode: 72,
		MinorUnit:   2,
	}
	currency23 = &Currency{
		Name:        "Belarussian Ruble",
		Code:        CurrencyCodeBYR,
		NumericCode: 974,
		MinorUnit:   0,
	}
	currency24 = &Currency{
		Name:        "Belize Dollar",
		Code:        CurrencyCodeBZD,
		NumericCode: 84,
		MinorUnit:   2,
	}
	currency25 = &Currency{
		Name:        "Canadian Dollar",
		Code:        CurrencyCodeCAD,
		NumericCode: 124,
		MinorUnit:   2,
	}
	currency26 = &Currency{
		Name:        "Swiss Franc",
		Code:        CurrencyCodeCHF,
		NumericCode: 756,
		MinorUnit:   2,
	}
	currency27 = &Currency{
		Name:        "Chilean Peso",
		Code:        CurrencyCodeCLP,
		NumericCode: 152,
		MinorUnit:   0,
	}
	currency28 = &Currency{
		Name:        "Yuan Renminbi",
		Code:        CurrencyCodeCNY,
		NumericCode: 156,
		MinorUnit:   2,
	}
	currency29 = &Currency{
		Name:        "Colombian Peso",
		Code:        CurrencyCodeCOP,
		NumericCode: 170,
		MinorUnit:   2,
	}
	currency30 = &Currency{
		Name:        "Costa Rican Colon",
		Code:        CurrencyCodeCRC,
		NumericCode: 188,
		MinorUnit:   2,
	}
	currency31 = &Currency{
		Name:        "Cuban Peso",
		Code:        CurrencyCodeCUP,
		NumericCode: 192,
		MinorUnit:   2,
	}
	currency32 = &Currency{
		Name:        "Cabo Verde Escudo",
		Code:        CurrencyCodeCVE,
		NumericCode: 132,
		MinorUnit:   2,
	}
	currency33 = &Currency{
		Name:        "Czech Koruna",
		Code:        CurrencyCodeCZK,
		NumericCode: 203,
		MinorUnit:   2,
	}
	currency34 = &Currency{
		Name:        "Djibouti Franc",
		Code:        CurrencyCodeDJF,
		NumericCode: 262,
		MinorUnit:   0,
	}
	currency35 = &Currency{
		Name:        "Danish Krone",
		Code:        CurrencyCodeDKK,
		NumericCode: 208,
		MinorUnit:   2,
	}
	currency36 = &Currency{
		Name:        "Dominican Peso",
		Code:        CurrencyCodeDOP,
		NumericCode: 214,
		MinorUnit:   2,
	}
	currency37 = &Currency{
		Name:        "Algerian Dinar",
		Code:        CurrencyCodeDZD,
		NumericCode: 12,
		MinorUnit:   2,
	}
	currency38 = &Currency{
		Name:        "Egyptian Pound",
		Code:        CurrencyCodeEGP,
		NumericCode: 818,
		MinorUnit:   2,
	}
	currency39 = &Currency{
		Name:        "Nakfa",
		Code:        CurrencyCodeERN,
		NumericCode: 232,
		MinorUnit:   2,
	}
	currency40 = &Currency{
		Name:        "Ethiopian Birr",
		Code:        CurrencyCodeETB,
		NumericCode: 230,
		MinorUnit:   2,
	}
	currency41 = &Currency{
		Name:        "Euro",
		Code:        CurrencyCodeEUR,
		NumericCode: 978,
		MinorUnit:   2,
	}
	currency42 = &Currency{
		Name:        "Fiji Dollar",
		Code:        CurrencyCodeFJD,
		NumericCode: 242,
		MinorUnit:   2,
	}
	currency43 = &Currency{
		Name:        "Falkland Islands Pound",
		Code:        CurrencyCodeFKP,
		NumericCode: 238,
		MinorUnit:   2,
	}
	currency44 = &Currency{
		Name:        "Pound Sterling",
		Code:        CurrencyCodeGBP,
		NumericCode: 826,
		MinorUnit:   2,
	}
	currency45 = &Currency{
		Name:        "Lari",
		Code:        CurrencyCodeGEL,
		NumericCode: 981,
		MinorUnit:   2,
	}
	currency46 = &Currency{
		Name:        "Ghana Cedi",
		Code:        CurrencyCodeGHS,
		NumericCode: 936,
		MinorUnit:   2,
	}
	currency47 = &Currency{
		Name:        "Gibraltar Pound",
		Code:        CurrencyCodeGIP,
		NumericCode: 292,
		MinorUnit:   2,
	}
	currency48 = &Currency{
		Name:        "Dalasi",
		Code:        CurrencyCodeGMD,
		NumericCode: 270,
		MinorUnit:   2,
	}
	currency49 = &Currency{
		Name:        "Guinea Franc",
		Code:        CurrencyCodeGNF,
		NumericCode: 324,
		MinorUnit:   0,
	}
	currency50 = &Currency{
		Name:        "Quetzal",
		Code:        CurrencyCodeGTQ,
		NumericCode: 320,
		MinorUnit:   2,
	}
	currency51 = &Currency{
		Name:        "Guyana Dollar",
		Code:        CurrencyCodeGYD,
		NumericCode: 328,
		MinorUnit:   2,
	}
	currency52 = &Currency{
		Name:        "Hong Kong Dollar",
		Code:        CurrencyCodeHKD,
		NumericCode: 344,
		MinorUnit:   2,
	}
	currency53 = &Currency{
		Name:        "Lempira",
		Code:        CurrencyCodeHNL,
		NumericCode: 340,
		MinorUnit:   2,
	}
	currency54 = &Currency{
		Name:        "Croatian Kuna",
		Code:        CurrencyCodeHRK,
		NumericCode: 191,
		MinorUnit:   2,
	}
	currency55 = &Currency{
		Name:        "Forint",
		Code:        CurrencyCodeHUF,
		NumericCode: 348,
		MinorUnit:   2,
	}
	currency56 = &Currency{
		Name:        "Rupiah",
		Code:        CurrencyCodeIDR,
		NumericCode: 360,
		MinorUnit:   2,
	}
	currency57 = &Currency{
		Name:        "New Israeli Sheqel",
		Code:        CurrencyCodeILS,
		NumericCode: 376,
		MinorUnit:   2,
	}
	currency58 = &Currency{
		Name:        "Indian Rupee",
		Code:        CurrencyCodeINR,
		NumericCode: 356,
		MinorUnit:   2,
	}
	currency59 = &Currency{
		Name:        "Iraqi Dinar",
		Code:        CurrencyCodeIQD,
		NumericCode: 368,
		MinorUnit:   3,
	}
	currency60 = &Currency{
		Name:        "Iranian Rial",
		Code:        CurrencyCodeIRR,
		NumericCode: 364,
		MinorUnit:   2,
	}
	currency61 = &Currency{
		Name:        "Iceland Krona",
		Code:        CurrencyCodeISK,
		NumericCode: 352,
		MinorUnit:   0,
	}
	currency62 = &Currency{
		Name:        "Jamaican Dollar",
		Code:        CurrencyCodeJMD,
		NumericCode: 388,
		MinorUnit:   2,
	}
	currency63 = &Currency{
		Name:        "Jordanian Dinar",
		Code:        CurrencyCodeJOD,
		NumericCode: 400,
		MinorUnit:   3,
	}
	currency64 = &Currency{
		Name:        "Yen",
		Code:        CurrencyCodeJPY,
		NumericCode: 392,
		MinorUnit:   0,
	}
	currency65 = &Currency{
		Name:        "Kenyan Shilling",
		Code:        CurrencyCodeKES,
		NumericCode: 404,
		MinorUnit:   2,
	}
	currency66 = &Currency{
		Name:        "Som",
		Code:        CurrencyCodeKGS,
		NumericCode: 417,
		MinorUnit:   2,
	}
	currency67 = &Currency{
		Name:        "Riel",
		Code:        CurrencyCodeKHR,
		NumericCode: 116,
		MinorUnit:   2,
	}
	currency68 = &Currency{
		Name:        "Comoro Franc",
		Code:        CurrencyCodeKMF,
		NumericCode: 174,
		MinorUnit:   0,
	}
	currency69 = &Currency{
		Name:        "North Korean Won",
		Code:        CurrencyCodeKPW,
		NumericCode: 408,
		MinorUnit:   2,
	}
	currency70 = &Currency{
		Name:        "Won",
		Code:        CurrencyCodeKRW,
		NumericCode: 410,
		MinorUnit:   0,
	}
	currency71 = &Currency{
		Name:        "Kuwaiti Dinar",
		Code:        CurrencyCodeKWD,
		NumericCode: 414,
		MinorUnit:   3,
	}
	currency72 = &Currency{
		Name:        "Cayman Islands Dollar",
		Code:        CurrencyCodeKYD,
		NumericCode: 136,
		MinorUnit:   2,
	}
	currency73 = &Currency{
		Name:        "Tenge",
		Code:        CurrencyCodeKZT,
		NumericCode: 398,
		MinorUnit:   2,
	}
	currency74 = &Currency{
		Name:        "Kip",
		Code:        CurrencyCodeLAK,
		NumericCode: 418,
		MinorUnit:   2,
	}
	currency75 = &Currency{
		Name:        "Lebanese Pound",
		Code:        CurrencyCodeLBP,
		NumericCode: 422,
		MinorUnit:   2,
	}
	currency76 = &Currency{
		Name:        "Sri Lanka Rupee",
		Code:        CurrencyCodeLKR,
		NumericCode: 144,
		MinorUnit:   2,
	}
	currency77 = &Currency{
		Name:        "Liberian Dollar",
		Code:        CurrencyCodeLRD,
		NumericCode: 430,
		MinorUnit:   2,
	}
	currency78 = &Currency{
		Name:        "Libyan Dinar",
		Code:        CurrencyCodeLYD,
		NumericCode: 434,
		MinorUnit:   3,
	}
	currency79 = &Currency{
		Name:        "Moroccan Dirham",
		Code:        CurrencyCodeMAD,
		NumericCode: 504,
		MinorUnit:   2,
	}
	currency80 = &Currency{
		Name:        "Moldovan Leu",
		Code:        CurrencyCodeMDL,
		NumericCode: 498,
		MinorUnit:   2,
	}
	currency81 = &Currency{
		Name:        "Malagasy Ariary",
		Code:        CurrencyCodeMGA,
		NumericCode: 969,
		MinorUnit:   2,
	}
	currency82 = &Currency{
		Name:        "Denar",
		Code:        CurrencyCodeMKD,
		NumericCode: 807,
		MinorUnit:   2,
	}
	currency83 = &Currency{
		Name:        "Kyat",
		Code:        CurrencyCodeMMK,
		NumericCode: 104,
		MinorUnit:   2,
	}
	currency84 = &Currency{
		Name:        "Tugrik",
		Code:        CurrencyCodeMNT,
		NumericCode: 496,
		MinorUnit:   2,
	}
	currency85 = &Currency{
		Name:        "Pataca",
		Code:        CurrencyCodeMOP,
		NumericCode: 446,
		MinorUnit:   2,
	}
	currency86 = &Currency{
		Name:        "Ouguiya",
		Code:        CurrencyCodeMRO,
		NumericCode: 478,
		MinorUnit:   2,
	}
	currency87 = &Currency{
		Name:        "Mauritius Rupee",
		Code:        CurrencyCodeMUR,
		NumericCode: 480,
		MinorUnit:   2,
	}
	currency88 = &Currency{
		Name:        "Rufiyaa",
		Code:        CurrencyCodeMVR,
		NumericCode: 462,
		MinorUnit:   2,
	}
	currency89 = &Currency{
		Name:        "Kwacha",
		Code:        CurrencyCodeMWK,
		NumericCode: 454,
		MinorUnit:   2,
	}
	currency90 = &Currency{
		Name:        "Mexican Peso",
		Code:        CurrencyCodeMXN,
		NumericCode: 484,
		MinorUnit:   2,
	}
	currency91 = &Currency{
		Name:        "Malaysian Ringgit",
		Code:        CurrencyCodeMYR,
		NumericCode: 458,
		MinorUnit:   2,
	}
	currency92 = &Currency{
		Name:        "Mozambique Metical",
		Code:        CurrencyCodeMZN,
		NumericCode: 943,
		MinorUnit:   2,
	}
	currency93 = &Currency{
		Name:        "Naira",
		Code:        CurrencyCodeNGN,
		NumericCode: 566,
		MinorUnit:   2,
	}
	currency94 = &Currency{
		Name:        "Cordoba Oro",
		Code:        CurrencyCodeNIO,
		NumericCode: 558,
		MinorUnit:   2,
	}
	currency95 = &Currency{
		Name:        "Norwegian Krone",
		Code:        CurrencyCodeNOK,
		NumericCode: 578,
		MinorUnit:   2,
	}
	currency96 = &Currency{
		Name:        "Nepalese Rupee",
		Code:        CurrencyCodeNPR,
		NumericCode: 524,
		MinorUnit:   2,
	}
	currency97 = &Currency{
		Name:        "New Zealand Dollar",
		Code:        CurrencyCodeNZD,
		NumericCode: 554,
		MinorUnit:   2,
	}
	currency98 = &Currency{
		Name:        "Rial Omani",
		Code:        CurrencyCodeOMR,
		NumericCode: 512,
		MinorUnit:   3,
	}
	currency99 = &Currency{
		Name:        "Nuevo Sol",
		Code:        CurrencyCodePEN,
		NumericCode: 604,
		MinorUnit:   2,
	}
	currency100 = &Currency{
		Name:        "Kina",
		Code:        CurrencyCodePGK,
		NumericCode: 598,
		MinorUnit:   2,
	}
	currency101 = &Currency{
		Name:        "Philippine Peso",
		Code:        CurrencyCodePHP,
		NumericCode: 608,
		MinorUnit:   2,
	}
	currency102 = &Currency{
		Name:        "Pakistan Rupee",
		Code:        CurrencyCodePKR,
		NumericCode: 586,
		MinorUnit:   2,
	}
	currency103 = &Currency{
		Name:        "Zloty",
		Code:        CurrencyCodePLN,
		NumericCode: 985,
		MinorUnit:   2,
	}
	currency104 = &Currency{
		Name:        "Guarani",
		Code:        CurrencyCodePYG,
		NumericCode: 600,
		MinorUnit:   0,
	}
	currency105 = &Currency{
		Name:        "Qatari Rial",
		Code:        CurrencyCodeQAR,
		NumericCode: 634,
		MinorUnit:   2,
	}
	currency106 = &Currency{
		Name:        "New Romanian Leu",
		Code:        CurrencyCodeRON,
		NumericCode: 946,
		MinorUnit:   2,
	}
	currency107 = &Currency{
		Name:        "Serbian Dinar",
		Code:        CurrencyCodeRSD,
		NumericCode: 941,
		MinorUnit:   2,
	}
	currency108 = &Currency{
		Name:        "Russian Ruble",
		Code:        CurrencyCodeRUB,
		NumericCode: 643,
		MinorUnit:   2,
	}
	currency109 = &Currency{
		Name:        "Rwanda Franc",
		Code:        CurrencyCodeRWF,
		NumericCode: 646,
		MinorUnit:   0,
	}
	currency110 = &Currency{
		Name:        "Saudi Riyal",
		Code:        CurrencyCodeSAR,
		NumericCode: 682,
		MinorUnit:   2,
	}
	currency111 = &Currency{
		Name:        "Solomon Islands Dollar",
		Code:        CurrencyCodeSBD,
		NumericCode: 90,
		MinorUnit:   2,
	}
	currency112 = &Currency{
		Name:        "Seychelles Rupee",
		Code:        CurrencyCodeSCR,
		NumericCode: 690,
		MinorUnit:   2,
	}
	currency113 = &Currency{
		Name:        "Sudanese Pound",
		Code:        CurrencyCodeSDG,
		NumericCode: 938,
		MinorUnit:   2,
	}
	currency114 = &Currency{
		Name:        "Swedish Krona",
		Code:        CurrencyCodeSEK,
		NumericCode: 752,
		MinorUnit:   2,
	}
	currency115 = &Currency{
		Name:        "Singapore Dollar",
		Code:        CurrencyCodeSGD,
		NumericCode: 702,
		MinorUnit:   2,
	}
	currency116 = &Currency{
		Name:        "Saint Helena Pound",
		Code:        CurrencyCodeSHP,
		NumericCode: 654,
		MinorUnit:   2,
	}
	currency117 = &Currency{
		Name:        "Leone",
		Code:        CurrencyCodeSLL,
		NumericCode: 694,
		MinorUnit:   2,
	}
	currency118 = &Currency{
		Name:        "Somali Shilling",
		Code:        CurrencyCodeSOS,
		NumericCode: 706,
		MinorUnit:   2,
	}
	currency119 = &Currency{
		Name:        "Surinam Dollar",
		Code:        CurrencyCodeSRD,
		NumericCode: 968,
		MinorUnit:   2,
	}
	currency120 = &Currency{
		Name:        "South Sudanese Pound",
		Code:        CurrencyCodeSSP,
		NumericCode: 728,
		MinorUnit:   2,
	}
	currency121 = &Currency{
		Name:        "Dobra",
		Code:        CurrencyCodeSTD,
		NumericCode: 678,
		MinorUnit:   2,
	}
	currency122 = &Currency{
		Name:        "Syrian Pound",
		Code:        CurrencyCodeSYP,
		NumericCode: 760,
		MinorUnit:   2,
	}
	currency123 = &Currency{
		Name:        "Lilangeni",
		Code:        CurrencyCodeSZL,
		NumericCode: 748,
		MinorUnit:   2,
	}
	currency124 = &Currency{
		Name:        "Baht",
		Code:        CurrencyCodeTHB,
		NumericCode: 764,
		MinorUnit:   2,
	}
	currency125 = &Currency{
		Name:        "Somoni",
		Code:        CurrencyCodeTJS,
		NumericCode: 972,
		MinorUnit:   2,
	}
	currency126 = &Currency{
		Name:        "Turkmenistan New Manat",
		Code:        CurrencyCodeTMT,
		NumericCode: 934,
		MinorUnit:   2,
	}
	currency127 = &Currency{
		Name:        "Tunisian Dinar",
		Code:        CurrencyCodeTND,
		NumericCode: 788,
		MinorUnit:   3,
	}
	currency128 = &Currency{
		Name:        "Pa’anga",
		Code:        CurrencyCodeTOP,
		NumericCode: 776,
		MinorUnit:   2,
	}
	currency129 = &Currency{
		Name:        "Turkish Lira",
		Code:        CurrencyCodeTRY,
		NumericCode: 949,
		MinorUnit:   2,
	}
	currency130 = &Currency{
		Name:        "Trinidad and Tobago Dollar",
		Code:        CurrencyCodeTTD,
		NumericCode: 780,
		MinorUnit:   2,
	}
	currency131 = &Currency{
		Name:        "New Taiwan Dollar",
		Code:        CurrencyCodeTWD,
		NumericCode: 901,
		MinorUnit:   2,
	}
	currency132 = &Currency{
		Name:        "Tanzanian Shilling",
		Code:        CurrencyCodeTZS,
		NumericCode: 834,
		MinorUnit:   2,
	}
	currency133 = &Currency{
		Name:        "Hryvnia",
		Code:        CurrencyCodeUAH,
		NumericCode: 980,
		MinorUnit:   2,
	}
	currency134 = &Currency{
		Name:        "Uganda Shilling",
		Code:        CurrencyCodeUGX,
		NumericCode: 800,
		MinorUnit:   0,
	}
	currency135 = &Currency{
		Name:        "US Dollar",
		Code:        CurrencyCodeUSD,
		NumericCode: 840,
		MinorUnit:   2,
	}
	currency136 = &Currency{
		Name:        "Peso Uruguayo",
		Code:        CurrencyCodeUYU,
		NumericCode: 858,
		MinorUnit:   2,
	}
	currency137 = &Currency{
		Name:        "Uzbekistan Sum",
		Code:        CurrencyCodeUZS,
		NumericCode: 860,
		MinorUnit:   2,
	}
	currency138 = &Currency{
		Name:        "Bolivar",
		Code:        CurrencyCodeVEF,
		NumericCode: 937,
		MinorUnit:   2,
	}
	currency139 = &Currency{
		Name:        "Dong",
		Code:        CurrencyCodeVND,
		NumericCode: 704,
		MinorUnit:   0,
	}
	currency140 = &Currency{
		Name:        "Vatu",
		Code:        CurrencyCodeVUV,
		NumericCode: 548,
		MinorUnit:   0,
	}
	currency141 = &Currency{
		Name:        "Tala",
		Code:        CurrencyCodeWST,
		NumericCode: 882,
		MinorUnit:   2,
	}
	currency142 = &Currency{
		Name:        "CFA Franc BEAC",
		Code:        CurrencyCodeXAF,
		NumericCode: 950,
		MinorUnit:   0,
	}
	currency143 = &Currency{
		Name:        "East Caribbean Dollar",
		Code:        CurrencyCodeXCD,
		NumericCode: 951,
		MinorUnit:   2,
	}
	currency144 = &Currency{
		Name:        "CFA Franc BCEAO",
		Code:        CurrencyCodeXOF,
		NumericCode: 952,
		MinorUnit:   0,
	}
	currency145 = &Currency{
		Name:        "CFP Franc",
		Code:        CurrencyCodeXPF,
		NumericCode: 953,
		MinorUnit:   0,
	}
	currency146 = &Currency{
		Name:        "Yemeni Rial",
		Code:        CurrencyCodeYER,
		NumericCode: 886,
		MinorUnit:   2,
	}
	currency147 = &Currency{
		Name:        "Rand",
		Code:        CurrencyCodeZAR,
		NumericCode: 710,
		MinorUnit:   2,
	}
	currency148 = &Currency{
		Name:        "Zambian Kwacha",
		Code:        CurrencyCodeZMW,
		NumericCode: 967,
		MinorUnit:   2,
	}
	currency149 = &Currency{
		Name:        "Zimbabwe Dollar",
		Code:        CurrencyCodeZWL,
		NumericCode: 932,
		MinorUnit:   2,
	}
)

// CurrencyCodeSimpleValueOf returns the value of the simple string s.
func CurrencyCodeSimpleValueOf(s string) (CurrencyCode, error) {
	value, ok := SimpleStringToCurrencyCode[strings.ToUpper(s)]
	if !ok {
		return CurrencyCode_CURRENCY_CODE_NONE, fmt.Errorf("pbmoney: no CurrencyCode for %s", s)
	}
	return value, nil
}

// SimpleString returns the simple string.
func (c CurrencyCode) SimpleString() string {
	simpleString, ok := CurrencyCodeToSimpleString[c]
	if !ok {
		return ""
	}
	return simpleString
}

// Currency returns the associated Currency, or nil if no Currency is known.
func (c CurrencyCode) Currency() *Currency {
	currency, ok := CurrencyCodeToCurrency[c]
	if !ok {
		return nil
	}
	return currency
}

func (c CurrencyCode) minorMultiplier() int64 {
	// TODO(pedge): will panic if c.Currency() == nil
	return int64(math.Pow(10, float64(6-c.Currency().MinorUnit)))
}

// NewMoney returns a new Money for the given CurrencyCode and valueMicros.
func NewMoney(currencyCode CurrencyCode, valueMicros int64) *Money {
	return &Money{
		CurrencyCode: currencyCode,
		ValueMicros:  valueMicros,
	}
}

// NewMoneyUSD returns a new Money for USD for the given valueMicros.
func NewMoneyUSD(valueMicros int64) *Money {
	return NewMoney(CurrencyCodeUSD, valueMicros)
}

// NewMoneyEUR returns a new Money for EUR for the given valueMicros.
func NewMoneyEUR(valueMicros int64) *Money {
	return NewMoney(CurrencyCodeEUR, valueMicros)
}

// NewMoneySimple returns a new Money for the given CurrencyCode and value.
//
// ValueMinorUnits will use the units of the CurrencyCode.
// If valueUnits is negative, valueMinorUnits will be converted to negative.
func NewMoneySimple(currencyCode CurrencyCode, valueUnits int64, valueMinorUnits int64) *Money {
	if valueUnits < 0 {
		valueMinorUnits = -valueMinorUnits
	}
	return &Money{
		CurrencyCode: currencyCode,
		ValueMicros:  unitsAndMicroPartToMicros(valueUnits, valueMinorUnits*currencyCode.minorMultiplier()),
	}
}

// NewMoneySimpleUSD returns a new Money for USD for the given and value.
//
// If valueDollars is negative, valueCents will be converted to negative.
func NewMoneySimpleUSD(valueDollars int64, valueCents int64) *Money {
	return NewMoneySimple(CurrencyCodeUSD, valueDollars, valueCents)
}

// NewMoneySimpleEUR returns a new Money for EUR for the given and value.
//
// If valueEuros is negative, valueCents will be converted to negative.
func NewMoneySimpleEUR(valueEuros int64, valueCents int64) *Money {
	return NewMoneySimple(CurrencyCodeEUR, valueEuros, valueCents)
}

// NewMoneyFloat returns a new Money for the given CurrencyCode and float value.
func NewMoneyFloat(currencyCode CurrencyCode, value float64) *Money {
	return &Money{
		CurrencyCode: currencyCode,
		ValueMicros:  floatUnitsToMicros(value),
	}
}

// NewMoneyFloatUSD returns a new Money for USD for the given and value.
func NewMoneyFloatUSD(valueDollars float64) *Money {
	return NewMoneyFloat(CurrencyCodeUSD, valueDollars)
}

// NewMoneyFloatEUR returns a new Money for EUR for the given and value.
func NewMoneyFloatEUR(valueEuros float64) *Money {
	return NewMoneyFloat(CurrencyCodeEUR, valueEuros)
}

// GoogleMoneyToMoney converts a google.type.Money to Money.
func GoogleMoneyToMoney(googleMoney *google_type.Money) (*Money, error) {
	currencyCode, err := CurrencyCodeSimpleValueOf(googleMoney.CurrencyCode)
	if err != nil {
		return nil, err
	}
	return &Money{
		CurrencyCode: currencyCode,
		ValueMicros:  unitsAndNanoPartToMicros(googleMoney.Units, googleMoney.Nanos),
	}, nil
}

// ToGoogleMoney converts the given Money into a google.type.Money
func (m *Money) ToGoogleMoney() *google_type.Money {
	units, nanoPart := microsToUnitsAndNanoPart(m.ValueMicros)
	return &google_type.Money{
		CurrencyCode: m.CurrencyCode.SimpleString(),
		Units:        units,
		Nanos:        nanoPart,
	}
}

// LessThan returns true if m is less than j, or error if m and j's CurrencyCodes are not the same.
func (m *Money) LessThan(j *Money) (bool, error) {
	if j == nil {
		return false, nil
	}
	if m == nil {
		return true, nil
	}
	if m.CurrencyCode != j.CurrencyCode {
		return false, fmt.Errorf("pbmoney: currency codes %s and %s do not match", m.CurrencyCode.SimpleString(), j.CurrencyCode.SimpleString())
	}
	return m.ValueMicros < j.ValueMicros, nil
}

// IsZero returns true if ValueMicros == 0.
func (m *Money) IsZero() bool {
	return m.ValueMicros == 0
}

// Units returns the units-only part of the value.
func (m *Money) Units() int64 {
	units, _ := microsToUnitsAndMicroPart(m.ValueMicros)
	return units
}

// Float returns the value of Money as a float, not in micros.
func (m *Money) Float() float64 {
	return microsToFloat(m.ValueMicros)
}

// SimpleString returns the simple string for the Money.
func (m *Money) SimpleString() string {
	if m == nil {
		return ""
	}
	units, microPart := microsToUnitsAndMicroPart(m.ValueMicros)
	if microPart < 0 {
		microPart = -microPart
	}
	return fmt.Sprintf("%d.%06d", units, microPart)
}

// MoneyMathable is the interface needed to perform money math operations.
type MoneyMathable interface {
	GetCurrencyCode() CurrencyCode
	GetValueMicros() int64
	Errors() []error
}

// MoneyMather performs math operations on Money.
type MoneyMather interface {
	MoneyMathable
	Plus(MoneyMathable) MoneyMather
	Minus(MoneyMathable) MoneyMather
	Times(MoneyMathable) MoneyMather
	Div(MoneyMathable) MoneyMather
	Min(MoneyMathable) MoneyMather
	Abs() MoneyMather
	PlusInt(int64) MoneyMather
	MinusInt(int64) MoneyMather
	TimesInt(int64) MoneyMather
	DivInt(int64) MoneyMather
	PlusFloat(float64) MoneyMather
	MinusFloat(float64) MoneyMather
	TimesFloat(float64) MoneyMather
	DivFloat(float64) MoneyMather
	Result() (*Money, error)
}

// Plus does the MoneyMather Plus operation.
func (m *Money) Plus(moneyMathable MoneyMathable) MoneyMather {
	return newMoneyMather(m).Plus(moneyMathable)
}

// Minus does the MoneyMather Minus operation.
func (m *Money) Minus(moneyMathable MoneyMathable) MoneyMather {
	return newMoneyMather(m).Minus(moneyMathable)
}

// Times does the MoneyMather Times operation.
func (m *Money) Times(moneyMathable MoneyMathable) MoneyMather {
	return newMoneyMather(m).Times(moneyMathable)
}

// Div does the MoneyMather Div operation.
func (m *Money) Div(moneyMathable MoneyMathable) MoneyMather {
	return newMoneyMather(m).Div(moneyMathable)
}

// Min does the MoneyMather Min operation.
func (m *Money) Min(moneyMathable MoneyMathable) MoneyMather {
	return newMoneyMather(m).Min(moneyMathable)
}

// Abs does the MoneyMather Abs operation.
func (m *Money) Abs() MoneyMather {
	return newMoneyMather(m).Abs()
}

// PlusInt does the MoneyMather Plus operation.
func (m *Money) PlusInt(value int64) MoneyMather {
	return newMoneyMather(m).PlusInt(value)
}

// MinusInt does the MoneyMather Minus operation.
func (m *Money) MinusInt(value int64) MoneyMather {
	return newMoneyMather(m).MinusInt(value)
}

// TimesInt does the MoneyMather Times operation.
func (m *Money) TimesInt(value int64) MoneyMather {
	return newMoneyMather(m).TimesInt(value)
}

// DivInt does the MoneyMather Div operation.
func (m *Money) DivInt(value int64) MoneyMather {
	return newMoneyMather(m).DivInt(value)
}

// PlusFloat does the MoneyMather Plus operation.
func (m *Money) PlusFloat(value float64) MoneyMather {
	return newMoneyMather(m).PlusFloat(value)
}

// MinusFloat does the MoneyMather Minus operation.
func (m *Money) MinusFloat(value float64) MoneyMather {
	return newMoneyMather(m).MinusFloat(value)
}

// TimesFloat does the MoneyMather Times operation.
func (m *Money) TimesFloat(value float64) MoneyMather {
	return newMoneyMather(m).TimesFloat(value)
}

// DivFloat does the MoneyMather Div operation.
func (m *Money) DivFloat(value float64) MoneyMather {
	return newMoneyMather(m).DivFloat(value)
}

// GetCurrencyCode returns the CurrencyCode.
func (m *Money) GetCurrencyCode() CurrencyCode {
	return m.CurrencyCode
}

// GetValueMicros returns the ValueMicros.
func (m *Money) GetValueMicros() int64 {
	return m.ValueMicros
}

// Errors returns nil.
func (m *Money) Errors() []error {
	return nil
}

type moneyMather struct {
	cc   CurrencyCode
	vm   int64
	errs []error
}

func newMoneyMather(money *Money) *moneyMather {
	var errs []error
	if money.CurrencyCode == CurrencyCode_CURRENCY_CODE_NONE {
		errs = append(errs, fmt.Errorf("pbmoney: cannot use Money with CurrencyCode_CURRENCY_CODE_NONE"))
	}
	return &moneyMather{
		cc:   money.CurrencyCode,
		vm:   money.ValueMicros,
		errs: errs,
	}
}

func (m *moneyMather) Plus(moneyMathable MoneyMathable) MoneyMather {
	if !m.ok(moneyMathable) {
		return m
	}
	m.vm += moneyMathable.GetValueMicros()
	return m
}

func (m *moneyMather) Minus(moneyMathable MoneyMathable) MoneyMather {
	if !m.ok(moneyMathable) {
		return m
	}
	m.vm -= moneyMathable.GetValueMicros()
	return m
}

func (m *moneyMather) Times(moneyMathable MoneyMathable) MoneyMather {
	if !m.ok(moneyMathable) {
		return m
	}
	m.vm *= moneyMathable.GetValueMicros()
	return m
}

func (m *moneyMather) Div(moneyMathable MoneyMathable) MoneyMather {
	if !m.ok(moneyMathable) {
		return m
	}
	if moneyMathable.GetValueMicros() == 0 {
		m.errs = append(m.errs, fmt.Errorf("pbmoney: cannot divide by 0"))
		return m
	}
	m.vm /= moneyMathable.GetValueMicros()
	return m
}

func (m *moneyMather) Min(moneyMathable MoneyMathable) MoneyMather {
	if !m.ok(moneyMathable) {
		return m
	}
	if m.vm > moneyMathable.GetValueMicros() {
		m.vm = moneyMathable.GetValueMicros()
	}
	return m
}

func (m *moneyMather) Abs() MoneyMather {
	if m.vm < 0 {
		m.vm = -m.vm
	}
	return m
}

func (m *moneyMather) PlusInt(value int64) MoneyMather {
	m.vm += value
	return m
}

func (m *moneyMather) MinusInt(value int64) MoneyMather {
	m.vm -= value
	return m
}

func (m *moneyMather) TimesInt(value int64) MoneyMather {
	m.vm *= value
	return m
}

func (m *moneyMather) DivInt(value int64) MoneyMather {
	if value == 0 {
		m.errs = append(m.errs, fmt.Errorf("pbmoney: cannot divide by 0"))
		return m
	}
	m.vm /= value
	return m
}

func (m *moneyMather) PlusFloat(value float64) MoneyMather {
	m.vm += int64(value)
	return m
}

func (m *moneyMather) MinusFloat(value float64) MoneyMather {
	m.vm -= int64(value)
	return m
}

func (m *moneyMather) TimesFloat(value float64) MoneyMather {
	m.vm = int64(float64(m.vm) * value)
	return m
}

func (m *moneyMather) DivFloat(value float64) MoneyMather {
	if value == 0.0 {
		m.errs = append(m.errs, fmt.Errorf("pbmoney: cannot divide by 0"))
		return m
	}
	m.vm = int64(float64(m.vm) / value)
	return m
}

func (m *moneyMather) Result() (*Money, error) {
	if len(m.errs) > 0 {
		return nil, fmt.Errorf("%v", m.errs)
	}
	return &Money{
		CurrencyCode: m.cc,
		ValueMicros:  m.vm,
	}, nil
}

func (m *moneyMather) GetCurrencyCode() CurrencyCode {
	return m.cc
}

func (m *moneyMather) GetValueMicros() int64 {
	return m.vm
}

func (m *moneyMather) Errors() []error {
	return m.errs
}

func (m *moneyMather) ok(moneyMathable MoneyMathable) bool {
	errs := moneyMathable.Errors()
	if len(errs) > 0 {
		m.errs = append(m.errs, errs...)
		return false
	}
	if m.cc != moneyMathable.GetCurrencyCode() {
		m.errs = append(m.errs, fmt.Errorf("pbmoney: mismatched CurrencyCodes: %s %s", m.cc, moneyMathable.GetCurrencyCode()))
		return false
	}
	return true
}

func unitsAndMicroPartToMicros(units int64, micros int64) int64 {
	return unitsToMicros(units) + micros
}

func unitsAndNanoPartToMicros(units int64, nanos int32) int64 {
	return unitsToMicros(units) + int64(nanos/1000)
}

func microsToUnitsAndMicroPart(micros int64) (int64, int64) {
	return micros / 1000000, micros % 1000000
}

func microsToUnitsAndNanoPart(micros int64) (int64, int32) {
	return micros / 1000000, int32(micros%1000000) * 1000
}

func unitsToMicros(units int64) int64 {
	return units * 1000000
}

func floatUnitsToMicros(floatUnits float64) int64 {
	return int64(floatUnits * 1000000.0)
}

func microsToFloat(micros int64) float64 {
	return float64(micros) / 1000000.0
}
