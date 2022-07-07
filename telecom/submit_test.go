package telecom

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewSubmit(t *testing.T) {
	subs := NewSubmit([]string{"17600001111", "17700001111"}, Poem, MtOptions{atTime: time.Now().Add(time.Minute)})
	assert.True(t, len(subs) == 4)

	for i, sub := range subs {
		t.Logf("%+v", sub)
		if i < 3 {
			assert.True(t, int(sub.PacketLength) == 328)
			assert.True(t, int(sub.msgLength) == 140)
			assert.True(t, int(sub.msgLength) == len(sub.msgBytes))
		} else {
			assert.True(t, int(sub.msgLength) <= 140)
			assert.True(t, int(sub.PacketLength) > 147)
		}
	}
}

func TestSubmit_Decode(t *testing.T) {

}

func TestSubmit_Encode(t *testing.T) {
	subs := NewSubmit([]string{"17600001111", "17700001111"}, Poem, MtOptions{atTime: time.Now().Add(time.Minute)})
	assert.True(t, len(subs) == 4)

	for _, sub := range subs {
		t.Logf("%+v", sub)
		dt := sub.Encode()
		assert.True(t, int(sub.PacketLength) == len(dt))
		t.Logf("%v: %x", int(sub.PacketLength) == len(dt), dt)
	}
}

func TestSubmit_ToResponse(t *testing.T) {

}

const Poem = "将进酒\n" +
	"君不见黄河之水天上来，奔流到海不复回。\n" +
	"君不见高堂明镜悲白发，朝如青丝暮成雪。\n" +
	"人生得意须尽欢，莫使金樽空对月。\n" +
	"天生我材必有用，千金散尽还复来。\n" +
	"烹羊宰牛且为乐，会须一饮三百杯。\n" +
	"岑夫子，丹丘生，将进酒，杯莫停。\n" +
	"与君歌一曲，请君为我倾耳听。\n" +
	"钟鼓馔玉不足贵，但愿长醉不愿醒。\n" +
	"古来圣贤皆寂寞，惟有饮者留其名。\n" +
	"陈王昔时宴平乐，斗酒十千恣欢谑。\n" +
	"主人何为言少钱，径须沽取对君酌。\n" +
	"五花马、千金裘，呼儿将出换美酒，与尔同销万古愁。"
