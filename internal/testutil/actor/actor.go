package actor

import (
	"github.com/yaegaki/dotlive-schedule-server/model"
)

// Iori .
var Iori = model.Actor{
	ID:      "iori",
	Hashtag: "#ヤマトイオリ",
}

// Pino .
var Pino = model.Actor{
	ID:      "pino",
	Hashtag: "#カルロピノ",
}

// Suzu .
var Suzu = model.Actor{
	ID:      "suzu",
	Hashtag: "#神楽すず",
}

// Chieri .
var Chieri = model.Actor{
	ID:      "chieri",
	Hashtag: "#花京院ちえり",
}

// Iroha .
var Iroha = model.Actor{
	ID:      "iroha",
	Hashtag: "#金剛いろは",
}

// Futaba .
var Futaba = model.Actor{
	ID:      "futaba",
	Hashtag: "#北上双葉",
}

// Mememe .
var Mememe = model.Actor{
	ID:      "mememe",
	Hashtag: "#もこ田めめめ",
}

// Siro .
var Siro = model.Actor{
	ID:      "siro",
	Hashtag: "#シロ生放送",
}

// Milk .
var Milk = model.Actor{
	ID:      "milk",
	Hashtag: "#メリーミルク",
}

// All .
var All = []model.Actor{
	Iori,
	Pino,
	Suzu,
	Chieri,
	Iroha,
	Futaba,
	Mememe,
	Siro,
	Milk,
}
