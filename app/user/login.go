package user

import "github.com/rpccloud/rpc"

var UserService = rpc.NewService().
	Reply("CreateUserByPhone", createUserByPhone).
	Reply("CreateUserByEmail", createUserByEmail).
	Reply("DeleteUser", deleteUser).
	Reply("AddPhone", addPhone).
	Reply("AddEmail", addEmail).
	Reply("DeleteItem", deleteItem).
	Reply("AuthItemLogin", authItemLogin)

func createUserByPhone(
	ctx rpc.Context,
	phone string,
	smsCode string,
) rpc.Return {
	return ctx.OK("createUserByPhone")
}

func createUserByEmail(
	ctx rpc.Context,
	email string,
) rpc.Return {
	return ctx.OK("createUserByEmail")
}

func deleteUser(
	ctx rpc.Context,
	security string,
	userID uint64,
) rpc.Return {
	return ctx.OK("deleteUser")
}

func addPhone(
	ctx rpc.Context,
	security string,
	userID uint64,
	phone string,
	smsCode string,
) rpc.Return {
	return ctx.OK("addPhone")
}

func addEmail(
	ctx rpc.Context,
	security string,
	userID uint64,
	email string,
) rpc.Return {
	return ctx.OK("addEmail")
}

func deleteItem(
	ctx rpc.Context,
	security string,
	authID uint64,
) rpc.Return {
	return ctx.OK("")
}

func authItemLogin(
	ctx rpc.Context,
	security string,
	authID uint64,
	canLogin bool,
) rpc.Return {
	return ctx.OK("authLogin")
}
