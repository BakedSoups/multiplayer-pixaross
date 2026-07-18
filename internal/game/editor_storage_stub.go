//go:build !js

package game

func saveCommunityProfile(string) bool { return false }

func loadCommunityProfile() string { return "" }

func communityAccountLabel() string { return "Sign in" }

func communitySignedIn() bool { return false }

func saveCommunityData(string) bool { return false }

func loadCommunityData() string { return "" }

func requestCommunityImport() bool { return false }

func takeCommunityImport() string { return "" }

func requestCommunitySignIn(string) bool { return false }

func requestCommunitySignOut() bool { return false }

func requestCommunityGoogleSignIn() bool { return false }

func requestCommunityPublish(string, bool, bool) bool { return false }

func requestCommunityPackPublish(string) bool { return false }

func takeCommunityResult() string { return "" }

func takeCommunityPublishedID() string { return "" }

func takeCommunityPublishedPackID() string { return "" }

func requestCommunityCatalog(string) bool { return false }

func takeCommunityCatalog() string { return "" }

func syncCommunityDraft(string) {}

func requestCommunityCloudDrafts() bool { return false }

func takeCommunityCloudDrafts() string { return "" }

func requestCommunityCreators() bool { return false }

func takeCommunityCreators() string { return "" }

func syncCommunityProfile(string) {}

func requestCommunityGallery(string, string) bool { return false }

func takeCommunityGallery() string { return "" }

func requestCommunityPublished() bool { return false }

func takeCommunityPublished() string { return "" }

func unpublishCommunityItem(string, string) bool { return false }

func toggleCommunityLike(string, string) bool { return false }

func promoteCommunityItem(string, string) bool { return false }

func saveEditorPack(string) bool {
	return false
}

func loadEditorPack() string {
	return ""
}

func exportEditorImage(string, string) bool {
	return false
}

func requestEditorImageImport(int) bool {
	return false
}

func takeEditorImageImport() string {
	return ""
}

func requestEditorColorPicker(string) bool {
	return false
}

func takeEditorColorPicker() string {
	return ""
}

func requestEditorTitle(string) bool { return false }

func takeEditorTitle() string { return "" }

func requestEditorPackImport() bool {
	return false
}

func takeEditorPackImport() string {
	return ""
}

func communityFetchStatus() string {
	return "Community is available in the web build"
}
