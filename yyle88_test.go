package yyle88_test

import (
	"fmt"
	"math/rand/v2"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yyle88/done"
	"github.com/yyle88/must"
	"github.com/yyle88/osexistpath/osmustexist"
	"github.com/yyle88/runpath"
	"github.com/yyle88/yyle88"
	"github.com/yyle88/yyle88/internal/utils"
)

const username = "yyle88"

var mutexRewriteFp sync.Mutex //write file one by one

type DocGenParam struct {
	shortName string
	titleLine string
	repoTitle string
}

func TestGenMarkdown(t *testing.T) {
	GenMarkdownTable(t, &DocGenParam{
		shortName: "README.md",
		titleLine: "| **RepoName** | **Description** |",
		repoTitle: "repo",
	})
}

func TestGenMarkdownZhHans(t *testing.T) {
	GenMarkdownTable(t, &DocGenParam{
		shortName: "README.zh.md",
		titleLine: "| 项目名称 | 项目描述 |",
		repoTitle: "项目",
	})
}

func TestGenMarkdownJapanese(t *testing.T) {
	GenMarkdownTable(t, &DocGenParam{
		shortName: "README.ja.md",
		titleLine: "| **リポ名** | **説明** |",
		repoTitle: "リポ",
	})
}

func TestGenMarkdownZhHant(t *testing.T) {
	GenMarkdownTable(t, &DocGenParam{
		shortName: "README.zh-Hant.md",
		titleLine: "| 倉庫名稱 | 描述 |",
		repoTitle: "倉庫",
	})
}

func TestGenMarkdownRussian(t *testing.T) {
	GenMarkdownTable(t, &DocGenParam{
		shortName: "README.ru.md",
		titleLine: "| Название проекта | Описание проекта |",
		repoTitle: "Проект",
	})
}

func GenMarkdownTable(t *testing.T, arg *DocGenParam) {
	mutexRewriteFp.Lock()
	defer mutexRewriteFp.Unlock()

	repos := fetchRepos()
	require.NotEmpty(t, repos)

	ptx := utils.NewPTX()

	cardThemes := utils.GetReadmeCardThemes()
	require.NotEmpty(t, cardThemes)

	subRepos, repos := splitRepos(repos, 5)
	for _, repo := range subRepos {
		const templateLine = "[![Readme Card](https://github-readme-stats.vercel.app/api/pin/?username={{ username }}&repo={{ repo_name }}&theme={{ card_theme }})]({{ repo_link }})"

		rep := strings.NewReplacer(
			"{{ username }}", username,
			"{{ repo_name }}", repo.Name,
			"{{ card_theme }}", cardThemes[rand.IntN(len(cardThemes))],
			"{{ repo_link }}", repo.Link,
		)

		ptx.Println(rep.Replace(templateLine))
		ptx.Println()
	}

	colors := utils.GetBadgeColors()
	require.NotEmpty(t, colors)

	rand.Shuffle(len(colors), func(i, j int) {
		colors[i], colors[j] = colors[j], colors[i]
	})

	subRepos, repos = splitRepos(repos, 5)
	if len(subRepos) > 0 {
		ptx.Println()
		ptx.Println(arg.titleLine)
		ptx.Println("|--------|--------|")
		for _, repo := range subRepos {
			ptx.Println(fmt.Sprintf("| %s | %s |", makeBadge(repo, colors[rand.IntN(len(colors))]), strings.ReplaceAll(repo.Desc, "|", "-")))
		}
		ptx.Println()
	}

	if len(repos) > 0 {
		ptx.Println()
		const stepLimit = 4
		ptx.Println("|" + repeatString(" "+arg.repoTitle+" |", stepLimit))
		ptx.Println("|" + repeatString(" :--: |", stepLimit))
		for start := 0; start < len(repos); start += stepLimit {
			ptx.Print("|")
			for num := 0; num < stepLimit; num++ {
				if idx := start + num; idx < len(repos) {
					ptx.Print(makeBadge(repos[idx], colors[idx%len(colors)]), " | ")
				} else {
					ptx.Print("-", " | ")
				}
			}
			ptx.Println()
		}
		ptx.Println()
	}

	stb := ptx.String()
	t.Log(stb)

	path := osmustexist.PATH(runpath.PARENT.Join(arg.shortName))
	t.Log(path)

	text := string(done.VAE(os.ReadFile(path)).Nice())
	t.Log(text)

	contentLines := strings.Split(text, "\n")
	sIdx := slices.Index(contentLines, "<!-- 这是一个注释，它不会在渲染时显示出来，这是项目列表的起始位置 -->")
	require.Positive(t, sIdx)
	eIdx := slices.Index(contentLines, "<!-- 这是一个注释，它不会在渲染时显示出来，这是项目列表的终止位置 -->")
	require.Positive(t, eIdx)

	require.Less(t, sIdx, eIdx)

	content := strings.Join(contentLines[:sIdx+1], "\n") + "\n" + "\n" +
		stb + "\n" +
		strings.Join(contentLines[eIdx:], "\n")
	t.Log(content)

	must.Done(os.WriteFile(path, []byte(content), 0666))
	t.Log("success")
}

var reposSingleton []*yyle88.Repo
var onceFetchRepos sync.Once

func fetchRepos() []*yyle88.Repo {
	onceFetchRepos.Do(func() {
		reposSingleton = done.VAE(yyle88.GetGithubRepos(username)).Nice()
	})
	return reposSingleton
}

func splitRepos(repos []*yyle88.Repo, subSize int) ([]*yyle88.Repo, []*yyle88.Repo) {
	idx := min(subSize, len(repos))
	return repos[:idx], repos[idx:]
}

func makeBadge(repo *yyle88.Repo, colorString string) string {
	return fmt.Sprintf("[![%s](https://img.shields.io/badge/%s-%s.svg?style=flat&logoColor=white)](%s)", repo.Name, repo.Name, url.QueryEscape(colorString), repo.Link)
}

func repeatString(s string, n int) string {
	var res string
	for i := 0; i < n; i++ {
		res += s
	}
	return res
}
