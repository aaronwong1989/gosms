package cmcc

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
	assert.Equal(t, d.destId, Conf.SrcId)
	assert.Equal(t, d.serviceId, Conf.ServiceId)
}

const Poem2 = "Will drink\n\nDon't you see the water of the Yellow River coming up from the sky, rushing to the sea and never returning.\n\nDon't you see the bright mirror of the high hall mourning white hair, like green silk in the morning and snow in the evening.\n\nWhen you are happy in life, don't make the golden cup empty to the moon.\n\nI'm born to be useful, but I'll come back after all the money is gone.\n\nCooking sheep and slaughtering cattle is fun, and you will have to drink 300 cups a day.\n\nMaster Cen, Dan Qiusheng, don't stop drinking.\n\nSing a song with you, please listen to it for me.\n\nBells, drums, and dishes are not expensive. I hope I'll be drunk for a long time and won't wake up.\n\nIn ancient times, saints and sages were lonely, and only drinkers kept their names.\n\nThe king of Chen used to enjoy banquets and drink ten thousand wine.\n\nWhy does the master say less money? He must sell and drink to you.\n\nFive flower horses, thousands of gold fur, hu er will exchange wine, and sell eternal sorrow with you."
