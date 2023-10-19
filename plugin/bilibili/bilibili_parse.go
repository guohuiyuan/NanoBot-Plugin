// Package bilibili bilibili视频解析
package bilibili

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	bz "github.com/FloatTech/AnimeAPI/bilibili"
	"github.com/FloatTech/floatbox/web"
	nano "github.com/fumiama/NanoBot"

	"github.com/FloatTech/NanoBot-Plugin/utils/ctxext"
	ctrl "github.com/FloatTech/zbpctrl"
)

var (
	searchVideo   = `bilibili.com\\?/video\\?/(?:av(\d+)|([bB][vV][0-9a-zA-Z]+))`
	searchVideoRe = regexp.MustCompile(searchVideo)
)

func init() {
	en := nano.Register("bilibiliparse", &ctrl.Options[*nano.Ctx]{
		DisableOnDefault: false,
		Brief:            "b站视频解析",
		Help:             "例:- bilibili.com/video/BV13B4y1x7pS",
	})
	en.OnMessageRegex(`((b23|acg).tv|bili2233.cn)/[0-9a-zA-Z]+`).SetBlock(true).Limit(ctxext.LimitByGroup).
		Handle(func(ctx *nano.Ctx) {
			url := ctx.State["regex_matched"].([]string)[0]
			realurl, err := bz.GetRealURL("https://" + url)
			if err != nil {
				_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
				return
			}
			switch {
			case searchVideoRe.MatchString(realurl):
				ctx.State["regex_matched"] = searchVideoRe.FindStringSubmatch(realurl)
				handleVideo(ctx)
			}
		})
	en.OnMessageRegex(searchVideo).SetBlock(true).Limit(ctxext.LimitByGroup).Handle(handleVideo)
}

func handleVideo(ctx *nano.Ctx) {
	id := ctx.State["regex_matched"].([]string)[1]
	if id == "" {
		id = ctx.State["regex_matched"].([]string)[2]
	}
	card, err := bz.GetVideoInfo(id)
	if err != nil {
		_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
		return
	}
	err = videoCard2msg(ctx, card)
	if err != nil {
		_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
		return
	}
	err = getVideoSummary(ctx, card)
	if err != nil {
		_, _ = ctx.SendPlainMessage(false, "ERROR: ", err)
		return
	}
}

// videoCard2msg 视频卡片转消息
func videoCard2msg(ctx *nano.Ctx, card bz.Card) (err error) {
	var mCard bz.MemberCard
	mCard, err = bz.GetMemberCard(card.Owner.Mid)
	if err != nil {
		return
	}
	file := card.Pic
	t := &strings.Builder{}
	t.WriteString("标题: ")
	t.WriteString(card.Title)
	t.WriteString("\n")
	if card.Rights.IsCooperation == 1 {
		for i := 0; i < len(card.Staff); i++ {
			t.WriteString(card.Staff[i].Title)
			t.WriteString(":")
			t.WriteString(card.Staff[i].Name)
			t.WriteString(" 粉丝: ")
			t.WriteString(bz.HumanNum(card.Staff[i].Follower))
			t.WriteString("\n")
		}
	} else {
		t.WriteString("UP主: ")
		t.WriteString(card.Owner.Name)
		t.WriteString(" 粉丝: ")
		t.WriteString(bz.HumanNum(mCard.Fans))
		t.WriteString("\n")
	}
	t.WriteString("UP主: ")
	t.WriteString(card.Owner.Name)
	t.WriteString(" 粉丝: ")
	t.WriteString(bz.HumanNum(mCard.Fans))
	t.WriteString("\n")

	t.WriteString("播放: ")
	t.WriteString(bz.HumanNum(card.Stat.View))
	t.WriteString(" 弹幕: ")
	t.WriteString(bz.HumanNum(card.Stat.Danmaku))
	t.WriteString("\n")

	t.WriteString("点赞: ")
	t.WriteString(bz.HumanNum(card.Stat.Like))
	t.WriteString(" 投币: ")
	t.WriteString(bz.HumanNum(card.Stat.Coin))
	t.WriteString("\n")

	t.WriteString("收藏: ")
	t.WriteString(bz.HumanNum(card.Stat.Favorite))
	t.WriteString(" 分享: ")
	t.WriteString(bz.HumanNum(card.Stat.Share))
	t.WriteString("\n")
	ctx.SendImage(file, false, t.String())
	return
}

// getVideoSummary AI视频总结
func getVideoSummary(ctx *nano.Ctx, card bz.Card) (err error) {
	var (
		data         []byte
		videoSummary bz.VideoSummary
	)
	data, err = web.GetData(bz.SignURL(fmt.Sprintf(bz.VideoSummaryURL, card.BvID, card.CID)))
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &videoSummary)
	t := &strings.Builder{}
	t.WriteString("已为你生成视频总结\n\n")
	t.WriteString(videoSummary.Data.ModelResult.Summary)
	t.WriteString("\n\n")
	for _, v := range videoSummary.Data.ModelResult.Outline {
		t.WriteString("● ")
		t.WriteString(v.Title)
		t.WriteString("\n")
		for _, p := range v.PartOutline {
			t.WriteString(strconv.Itoa(p.Timestamp / 60))
			t.WriteString(" ")
			t.WriteString(strconv.Itoa(p.Timestamp % 60))
			t.WriteString(" ")
			t.WriteString(p.Content)
			t.WriteString("\n")
		}
		t.WriteString("\n")
	}
	ctx.SendPlainMessage(false, t.String())
	return
}
