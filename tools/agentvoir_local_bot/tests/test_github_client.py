"""HTTP-mocked tests for the GitHub API client."""

from __future__ import annotations

from agentvoir_local_bot.github_client import GitHubClient


def test_list_candidate_issues_filters_claimed(httpx_mock, settings):
    """Only unclaimed, non-PR issues with the trigger label are returned."""
    httpx_mock.add_response(
        json=[
            {
                "number": 1,
                "title": "Ready",
                "body": "do it",
                "html_url": "https://github.com/acme/repo/issues/1",
                "labels": [{"name": "ai-code"}],
            },
            {
                "number": 2,
                "title": "Already claimed",
                "body": None,
                "html_url": "https://github.com/acme/repo/issues/2",
                "labels": [{"name": "ai-code"}, {"name": "ai-code-claimed"}],
            },
            {
                "number": 3,
                "title": "PR issue",
                "body": None,
                "html_url": "https://github.com/acme/repo/issues/3",
                "labels": [{"name": "ai-code"}],
                "pull_request": {"url": "https://api.github.com/repos/acme/repo/pulls/9"},
            },
        ]
    )

    with GitHubClient(settings) as client:
        issues = client.list_candidate_issues()

    assert len(issues) == 1
    assert issues[0].number == 1


def test_create_pull_request(httpx_mock, settings):
    """create_pull_request posts to the repo pulls endpoint and returns the response."""
    httpx_mock.add_response(
        method="POST",
        json={"html_url": "https://github.com/acme/repo/pull/10", "number": 10},
    )

    with GitHubClient(settings) as client:
        pr = client.create_pull_request(
            title="Fix thing",
            body="Closes #1",
            head="ai/issue-1-fix-thing",
            base="main",
        )

    assert pr["number"] == 10
    request = httpx_mock.get_request()
    assert request is not None
    assert request.url.path.endswith("/repos/acme/repo/pulls")


def test_claim_issue_replaces_trigger_label(httpx_mock, settings):
    """claim_issue removes ai-code and adds ai-code-claimed."""
    httpx_mock.add_response(method="DELETE")
    httpx_mock.add_response(method="POST")
    httpx_mock.add_response(method="POST")

    with GitHubClient(settings) as client:
        client.claim_issue(1)

    requests = httpx_mock.get_requests()
    assert requests[0].method == "DELETE"
    assert requests[0].url.path.endswith("/issues/1/labels/ai-code")
    assert requests[1].method == "POST"
    assert requests[1].url.path.endswith("/issues/1/labels")
    assert requests[1].content == b'{"labels":["ai-code-claimed"]}'
    assert requests[2].url.path.endswith("/issues/1/comments")
