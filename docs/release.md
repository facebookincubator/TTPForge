# Create New Release

This requires the [GitHub CLI](https://github.com/cli/cli#installation)
and [gh-changelog GitHub CLI extension](https://github.com/chelnak/gh-changelog).

Install changelog extension:

```bash
gh extension install chelnak/gh-changelog
```

Generate changelog:

```bash
NEXT_VERSION=v1.1.3
gh changelog new --next-version "${NEXT_VERSION}"
```

Create release:

```bash
gh release create "${NEXT_VERSION}" -F CHANGELOG.md
```
