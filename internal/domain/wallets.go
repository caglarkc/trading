package domain

import "github.com/google/uuid"

var WalletCodes = []string{"US-FAST", "US-EVENT", "US-PATIENT", "QQQ-FAST", "QQQ-PATIENT", "TQQQ-FAST", "TQQQ-PATIENT", "EWC-BALANCED"}

func WalletID(code string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("trading:wallet:"+code))
}
