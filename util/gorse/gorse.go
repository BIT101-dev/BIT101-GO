/**
* @author:YHCnb
* Package:
* @date:2023/10/18 21:17
* Description:
 */
package gorse

import (
	"context"
	"github.com/zhenghaoz/gorse/client"
)

var gorse *client.GorseClient

// Init 初始化
func Init() {
	gorse = client.NewGorseClient("http://127.0.0.1:8087", "api_key")

	gorse.InsertFeedback(context.Background(), []client.Feedback{
		{FeedbackType: "star", UserId: "bob", ItemId: "vuejs:vue", Timestamp: "2022-02-24"},
		{FeedbackType: "star", UserId: "bob", ItemId: "d3:d3", Timestamp: "2022-02-25"},
		{FeedbackType: "star", UserId: "bob", ItemId: "dogfalo:materialize", Timestamp: "2022-02-26"},
		{FeedbackType: "star", UserId: "bob", ItemId: "mozilla:pdf.js", Timestamp: "2022-02-27"},
		{FeedbackType: "star", UserId: "bob", ItemId: "moment:moment", Timestamp: "2022-02-28"},
	})

	// Get recommendation.
	gorse.GetRecommend(context.Background(), "bob", "", 10)
}
