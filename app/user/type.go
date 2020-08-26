package user

type MongoDBUser struct {
	ID         int64  `bson:"_id"`
	SecurityL1 string `bson:"security_l1"` // light user
	SecurityL2 string `bson:"security_l2"` // normal user
	SecurityL3 string `bson:"security_l3"` // admin user
}
