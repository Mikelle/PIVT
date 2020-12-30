///////////////////////////////////////////////////////////////////
// File containing the constants for the Cache Coin Smart Contract
///////////////////////////////////////////////////////////////////

package main

/*--------------------------------------------------
 SMART CONTRACT INDEXES
--------------------------------------------------*/

const IndexWallets = "WALLETS"
const IndexToken = "TOKEN"
const IndexCryptoList = "CRYPTO_LIST"
const IndexSocialTokenList = "SOCIAL_TOKEN_LIST"
const IndexFTPODList = "FT_POD_LIST"
const IndexNFTPODList = "NFT_POD_LIST"

const IndexFinancialScores = "SCORES"
const IndexBalances = "BALANCES"

const PRECISSION = 1e-8

/*--------------------------------------------------
 TOKEN TYPES
--------------------------------------------------*/
const CRYPTO_TOKEN = "CRYPTO"
const SOCIAL_TOKEN = "SOCIAL"
const FT_POD_TOKEN = "FTPOD"
const NFT_POD_TOKEN = "NFTPOD"

var TOKEN_TYPES = []string{
	CRYPTO_TOKEN, SOCIAL_TOKEN,
	FT_POD_TOKEN, NFT_POD_TOKEN}

/*--------------------------------------------------
 SYSTEM ROLES
--------------------------------------------------*/

const ADMIN_ROLE = "ADMIN"
const USER_ROLE = "USER"
const BUSINESS_ROLE = "BUSINESS"
const GUARANTOR_ROLE = "GUARANTOR"
const COURTMEMBER_ROLE = "COURT_MEMBER"
const EXCHANGE_ROLE = "EXCHANGE"

/*--------------------------------------------------
 SMART CONTRACT INVOKATIONS
--------------------------------------------------*/

const DATA_PROTOCOL_CHAINCODE = "DataProtocol"
const CHANNEL_NAME = "broadcast"

/*--------------------------------------------------
--------------------------------------------------*/
