package cmpp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDelivery(t *testing.T) {
	cases := []string{
		"hello world",
		"你好，世界。 hello world",
		"中华人民共和国",
		Poem,
		Poem2,
	}

	for _, msg := range cases {
		testcase(t, msg)
	}
}

func testcase(t *testing.T, msg string) {
	d := NewDelivery("17011110000", msg, "", "")
	t.Logf("%v", d)
	bts := d.Encode()
	t.Logf("len: %d, data: %x", len(bts), bts)
	assert.Equal(t, uint32(len(bts)), d.TotalLength)

	l := 0
	if len(msg) == len([]rune(msg)) {
		l += len(msg)
	} else {
		l += 2 * len([]rune(msg))
	}
	if d.msgFmt == 8 && len([]rune(msg)) > 70 {
		l = 140
	}
	if d.msgFmt != 8 && len([]rune(msg)) > 160 {
		l = 160
	}

	assert.Equal(t, d.msgLength, uint8(l))
	assert.Equal(t, d.destId, Conf.GetString("sms-display-no"))
	assert.Equal(t, d.serviceId, Conf.GetString("service-id"))
}

const Poem2 = "Will drink\n" +
	"Don't you see the water of the Yellow River coming up from the sky, rushing to the sea and never returning.\n" +
	"Don't you see the bright mirror of the high hall mourning white hair, like green silk in the morning and snow in the evening.\n" +
	"When you are happy in life, don't make the golden cup empty to the moon.\n" +
	"I'm born to be useful, but I'll come back after all the money is gone.\n" +
	"Cooking sheep and slaughtering cattle is fun, and you will have to drink 300 cups a day.\n" +
	"Master Cen, Dan Qiusheng, don't stop drinking.\n" +
	"Sing a song with you, please listen to it for me.\n" +
	"Bells, drums, and dishes are not expensive. I hope I'll be drunk for a long time and won't wake up.\n" +
	"In ancient times, saints and sages were lonely, and only drinkers kept their names.\n" +
	"The king of Chen used to enjoy banquets and drink ten thousand wine.\n" +
	"Why does the master say less money? He must sell and drink to you.\n" +
	"Five flower horses, thousands of gold fur, hu er will exchange wine, and sell eternal sorrow with you."
