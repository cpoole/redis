package redis_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	redis "github.com/go-redis/redis/v8"
)

type dataHash struct {
	Boolean bool   `redis:"boolean"`
	String  string `redis:"string"`
	Uint8   uint8  `redis:"uint8"`
	Int     int    `redis:"int"`
}

var _ = Describe("Cmd", func() {
	var client *redis.Client

	BeforeEach(func() {
		client = redis.NewClient(redisOptions())
		Expect(client.FlushDB(ctx).Err()).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(client.Close()).NotTo(HaveOccurred())
	})

	It("implements Stringer", func() {
		set := client.Set(ctx, "foo", "bar", 0)
		Expect(set.String()).To(Equal("set foo bar: OK"))

		get := client.Get(ctx, "foo")
		Expect(get.String()).To(Equal("get foo: bar"))
	})

	It("has val/err", func() {
		set := client.Set(ctx, "key", "hello", 0)
		Expect(set.Err()).NotTo(HaveOccurred())
		Expect(set.Val()).To(Equal("OK"))

		get := client.Get(ctx, "key")
		Expect(get.Err()).NotTo(HaveOccurred())
		Expect(get.Val()).To(Equal("hello"))

		Expect(set.Err()).NotTo(HaveOccurred())
		Expect(set.Val()).To(Equal("OK"))
	})

	It("has helpers", func() {
		set := client.Set(ctx, "key", "10", 0)
		Expect(set.Err()).NotTo(HaveOccurred())

		n, err := client.Get(ctx, "key").Int64()
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(int64(10)))

		un, err := client.Get(ctx, "key").Uint64()
		Expect(err).NotTo(HaveOccurred())
		Expect(un).To(Equal(uint64(10)))

		f, err := client.Get(ctx, "key").Float64()
		Expect(err).NotTo(HaveOccurred())
		Expect(f).To(Equal(float64(10)))
	})

	It("supports float32", func() {
		f := float32(66.97)

		err := client.Set(ctx, "float_key", f, 0).Err()
		Expect(err).NotTo(HaveOccurred())

		val, err := client.Get(ctx, "float_key").Float32()
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal(f))
	})

	It("supports time.Time", func() {
		tm := time.Date(2019, 1, 1, 9, 45, 10, 222125, time.UTC)

		err := client.Set(ctx, "time_key", tm, 0).Err()
		Expect(err).NotTo(HaveOccurred())

		s, err := client.Get(ctx, "time_key").Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(s).To(Equal("2019-01-01T09:45:10.000222125Z"))

		tm2, err := client.Get(ctx, "time_key").Time()
		Expect(err).NotTo(HaveOccurred())
		Expect(tm2).To(BeTemporally("==", tm))
	})

	It("does not stop scanning on first failure", func() {
		incorrectData := map[string]interface{}{
			"boolean": "-1", //this is the one that causes the error
			"string":  "foo",
			"uint8":   "1",
			"int":     "1",
		}

		client.HSet(ctx, "set_key", incorrectData)

		d := dataHash{}

		err := client.HGetAll(ctx, "set_key").Scan(&d)
		Expect(err).To(HaveOccurred())

		Expect(d).To(Equal(dataHash{
			Boolean: false,
			String:  "foo",
			Int:     1,
			Uint8:   1,
		}))

	})

	It("allows to set custom error", func() {
		e := errors.New("custom error")
		cmd := redis.Cmd{}
		cmd.SetErr(e)
		_, err := cmd.Result()
		Expect(err).To(Equal(e))
	})
})
