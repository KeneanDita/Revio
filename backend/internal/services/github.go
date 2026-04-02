package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/revio/backend/internal/models"
	"gorm.io/gorm"
)

const githubAPIBase = "https://api.github.com"

type GitHubService struct {
	db         *gorm.DB
	httpClient *http.Client
}

func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type GHRepository struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	Description string  `json:"description"`
	Private     bool    `json:"private"`
	HTMLURL     string  `json:"html_url"`
	Owner       GHOwner `json:"owner"`
}

type GHOwner struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type GHPullRequest struct {
	ID        int64      `json:"id"`
	Number    int        `json:"number"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	State     string     `json:"state"`
	HTMLURL   string     `json:"html_url"`
	MergedAt  *time.Time `json:"merged_at"`
	ClosedAt  *time.Time `json:"closed_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	User      struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
	Head struct {
		Ref string `json:"ref"`
	} `json:"head"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
	ChangedFiles int `json:"changed_files"`
	Commits      int `json:"commits"`
	Comments     int `json:"comments"`
}

type GHReview struct {
	ID          int64     `json:"id"`
	State       string    `json:"state"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	SubmittedAt time.Time `json:"submitted_at"`
	User        struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

type GHComment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	User      struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
}

func (s *GitHubService) GetRepository(ctx context.Context, token, fullName string) (*GHRepository, error) {
	url := fmt.Sprintf("%s/repos/%s", githubAPIBase, fullName)
	var repo GHRepository
	if err := s.get(ctx, token, url, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

func (s *GitHubService) ListUserRepositories(ctx context.Context, token string) ([]GHRepository, error) {
	url := fmt.Sprintf("%s/user/repos?per_page=100&sort=updated&type=all", githubAPIBase)
	var repos []GHRepository
	if err := s.get(ctx, token, url, &repos); err != nil {
		return nil, err
	}
	return repos, nil
}

func (s *GitHubService) SyncRepository(repoID uuid.UUID, token string) {
	ctx := context.Background()

	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", repoID).Error; err != nil {
		return
	}

	job := models.SyncJob{
		RepoID: repoID,
		Status: "running",
	}
	now := time.Now()
	job.StartedAt = &now
	s.db.Create(&job)

	if err := s.syncPullRequests(ctx, token, &repo); err != nil {
		errMsg := err.Error()
		job.Status = "failed"
		job.ErrorMsg = errMsg
	} else {
		job.Status = "completed"
		syncTime := time.Now()
		s.db.Model(&repo).Update("last_sync_at", syncTime)
	}

	finishedAt := time.Now()
	job.FinishedAt = &finishedAt
	s.db.Save(&job)
}

func (s *GitHubService) syncPullRequests(ctx context.Context, token string, repo *models.Repository) error {
	states := []string{"open", "closed"}

	for _, state := range states {
		page := 1
		for {
			url := fmt.Sprintf("%s/repos/%s/pulls?state=%s&per_page=100&page=%d",
				githubAPIBase, repo.FullName, state, page)

			var prs []GHPullRequest
			if err := s.get(ctx, token, url, &prs); err != nil {
				return fmt.Errorf("failed to fetch PRs: %w", err)
			}
			if len(prs) == 0 {
				break
			}

			for _, ghPR := range prs {
				s.upsertPullRequest(ctx, token, repo, &ghPR)
			}

			if len(prs) < 100 {
				break
			}
			page++
		}
	}

	return nil
}

func (s *GitHubService) upsertPullRequest(ctx context.Context, token string, repo *models.Repository, ghPR *GHPullRequest) {
	status := models.PRStatus(ghPR.State)
	if ghPR.MergedAt != nil {
		status = models.PRStatusMerged
	}

	var existing models.PullRequest
	isNew := s.db.Where("repo_id = ? AND github_id = ?", repo.ID, ghPR.ID).First(&existing).Error != nil

	pr := models.PullRequest{
		RepoID:          repo.ID,
		GitHubID:        ghPR.ID,
		Number:          ghPR.Number,
		Title:           ghPR.Title,
		Body:            ghPR.Body,
		Author:          ghPR.User.Login,
		AuthorAvatarURL: ghPR.User.AvatarURL,
		Status:          status,
		BaseBranch:      ghPR.Base.Ref,
		HeadBranch:      ghPR.Head.Ref,
		Additions:       ghPR.Additions,
		Deletions:       ghPR.Deletions,
		ChangedFiles:    ghPR.ChangedFiles,
		CommitCount:     ghPR.Commits,
		CommentCount:    ghPR.Comments,
		HTMLURL:         ghPR.HTMLURL,
		MergedAt:        ghPR.MergedAt,
		ClosedAt:        ghPR.ClosedAt,
		GithubCreatedAt: ghPR.CreatedAt,
		GithubUpdatedAt: ghPR.UpdatedAt,
	}

	if isNew {
		s.db.Create(&pr)
		s.syncReviews(ctx, token, repo, &pr, ghPR.Number)
	} else {
		pr.ID = existing.ID
		s.db.Save(&pr)
		s.syncReviews(ctx, token, repo, &pr, ghPR.Number)
	}
}

func (s *GitHubService) syncReviews(ctx context.Context, token string, repo *models.Repository, pr *models.PullRequest, prNumber int) {
	url := fmt.Sprintf("%s/repos/%s/pulls/%d/reviews?per_page=100",
		githubAPIBase, repo.FullName, prNumber)

	var ghReviews []GHReview
	if err := s.get(ctx, token, url, &ghReviews); err != nil {
		return
	}

	var firstReviewAt *time.Time

	for _, ghReview := range ghReviews {
		review := models.Review{
			PRID:           pr.ID,
			GitHubID:       ghReview.ID,
			Reviewer:       ghReview.User.Login,
			ReviewerAvatar: ghReview.User.AvatarURL,
			State:          models.ReviewState(ghReview.State),
			Body:           ghReview.Body,
			HTMLURL:        ghReview.HTMLURL,
			SubmittedAt:    ghReview.SubmittedAt,
		}

		s.db.Where(models.Review{GitHubID: review.GitHubID}).FirstOrCreate(&review)

		if firstReviewAt == nil || ghReview.SubmittedAt.Before(*firstReviewAt) {
			t := ghReview.SubmittedAt
			firstReviewAt = &t
		}
	}

	reviewCount := len(ghReviews)
	updates := map[string]interface{}{"review_count": reviewCount}
	if firstReviewAt != nil {
		updates["first_review_at"] = firstReviewAt
	}
	s.db.Model(pr).Updates(updates)
}

func (s *GitHubService) PostComment(ctx context.Context, token, owner, repo string, prNumber int, body string) (*GHComment, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", githubAPIBase, owner, repo, prNumber)

	payload := map[string]string{"body": body}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")
	req.Body = io.NopCloser(jsonReader(payloadBytes))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var comment GHComment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

func (s *GitHubService) get(ctx context.Context, token, url string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", url)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("GitHub API error: %d for %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

type bytesReader struct {
	data []byte
	pos  int
}

func jsonReader(b []byte) *bytesReader {
	return &bytesReader{data: b}
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
