package main



import (
	"log"
	"time"


	"PushServer/slog"
	"PushServer/redispool"
)

func tstSingle(pool *redispool.RedisPool) {

	cmd0 := []interface{}{"echo", "Hello world A0!"}
	rp := pool.CmdSingle("10.241.221.106:9600", cmd0)

	log.Println("tstSingle", rp)

}

func tst(pool *redispool.RedisPool) {

	cmd0 := []interface{}{"echo", "Hello world A0!"}
	cmd1 := []interface{}{"echo", "Hello world B0!"}
	cmd2 := []interface{}{"echo", "Hello world C0!"}

	//log.Println(reflect.TypeOf(cmd0[0]))

	mcmd := make(map[string][]interface{})
	mcmd["127.0.0.1:9600"] = cmd0
	mcmd["127.0.0.1:9601"] = cmd1
	mcmd["127.0.0.1:9602"] = cmd2

	log.Println(mcmd)
	rp := pool.Cmd(mcmd)



	log.Println("tst", rp)



}

func tst_base() {

	pool := redispool.NewRedisPool()
	log.Println(1, pool)

	go tst(pool)

	log.Println(3, pool)

	log.Println("--------------------")

	go tst(pool)

	log.Println(4, pool)

	log.Println("--------------------")

	for i := 0; i < 1000; i++ {
		go tstSingle(pool)
	}

	log.Println(5, pool)


	time.Sleep(time.Second * time.Duration(1))
	log.Println("--------------------")

	go tstSingle(pool)

	log.Println(6, pool)

	time.Sleep(time.Second * time.Duration(1))


}


func tst_redis_timeout() {

	pool := redispool.NewRedisPool()
	log.Println(1, pool)

	tstSingle(pool)

	log.Println(2, pool)

	// wait timeout
	time.Sleep(time.Second * time.Duration(12))

	tstSingle(pool)

	log.Println(3, pool)

}


func main() {
	slog.Init("")

	tst_redis_timeout()
}
