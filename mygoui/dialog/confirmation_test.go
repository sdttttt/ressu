package dialog

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestConfirmationShowsChinesePromptAndConfirms(t *testing.T) {
	confirmations := NewConfirmations(func() (string, string) {
		return "system", "small"
	})
	request := ConfirmationRequest{
		Title:        "取消订阅「测试源」？",
		Description:  "这个订阅源已保存的文章会一并删除，操作无法撤销。",
		ConfirmLabel: "取消订阅",
		CancelLabel:  "取消",
		Destructive:  true,
	}
	confirmation := &queuedConfirmation{
		ConfirmationRequest: request,
		done:                make(chan struct{}),
	}
	confirmations.enqueue(confirmation)

	tester := ui.NewTester(confirmations.Draw, 520, 420)
	tester.Frame()
	tester.Frame()
	for _, text := range []string{
		request.Title,
		request.Description,
		request.ConfirmLabel,
		request.CancelLabel,
	} {
		if !tester.HasText(text) {
			t.Errorf("confirmation is missing %q", text)
		}
	}

	tester.Click(request.ConfirmLabel)
	<-confirmation.done
	if !confirmation.result.accepted || confirmation.result.err != nil {
		t.Fatalf("confirmation result = %v, %v; want accepted", confirmation.result.accepted, confirmation.result.err)
	}
}

func TestConfirmationCancelDismissesWithoutRemoval(t *testing.T) {
	confirmations := NewConfirmations(func() (string, string) {
		return "system", "small"
	})
	confirmation := &queuedConfirmation{
		ConfirmationRequest: ConfirmationRequest{
			Title:        "取消订阅「测试源」？",
			Description:  "这个订阅源已保存的文章会一并删除，操作无法撤销。",
			ConfirmLabel: "取消订阅",
			CancelLabel:  "取消",
			Destructive:  true,
		},
		done: make(chan struct{}),
	}
	confirmations.enqueue(confirmation)

	tester := ui.NewTester(confirmations.Draw, 520, 420)
	tester.Frame()
	tester.Click(confirmation.CancelLabel)
	<-confirmation.done
	if confirmation.result.accepted || confirmation.result.err != nil {
		t.Fatalf("confirmation result = %v, %v; want canceled", confirmation.result.accepted, confirmation.result.err)
	}
}
