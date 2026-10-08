package io.github.maas3n.mattmux

/** SAF-independent Blu-ray metadata scanning, isolated from the DVD engine. */
internal data class BlurayDocumentEntry(
    val documentId: String,
    val displayName: String,
    val isDirectory: Boolean,
)
internal data class BluraySafPlaylist(
    val number: Int,
    val documentId: String,
    val navigation: BlurayPlaylistNavigation,
)
internal interface BlurayDocumentProvider {
    fun listChildren(parentDocumentId: String): List<BlurayDocumentEntry>
    fun readPlaylist(documentId: String): ByteArray
}

internal object BluraySafPlaylistCatalog {
    private fun uniqueDirectory(parent: String, name: String, provider: BlurayDocumentProvider): String {
        val matches = provider.listChildren(parent).filter {
            it.isDirectory && it.displayName.equals(name, ignoreCase = true)
        }
        require(matches.size == 1) { "Blu-ray folder " + name + " was not found or is ambiguous" }
        return matches.single().documentId
    }

    fun scan(root: BlurayDocumentEntry, provider: BlurayDocumentProvider): List<BluraySafPlaylist> {
        require(root.isDirectory) { "Select a Blu-ray folder, not an ISO file" }
        val bdmv = if (root.displayName.equals("BDMV", ignoreCase = true)) {
            root.documentId
        } else {
            uniqueDirectory(root.documentId, "BDMV", provider)
        }
        val playlistsFolder = uniqueDirectory(bdmv, "PLAYLIST", provider)
        val files = provider.listChildren(playlistsFolder)
            .filter { !it.isDirectory && it.displayName.matches(Regex("[0-9]{5}\\.mpls", RegexOption.IGNORE_CASE)) }
            .sortedWith(compareBy(String.CASE_INSENSITIVE_ORDER) { it.displayName })
        require(files.isNotEmpty()) { "Blu-ray PLAYLIST does not contain numbered MPLS files" }
        require(files.size <= 10_000) { "Too many Blu-ray playlists" }
        val used = hashSetOf<Int>()
        return files.map { entry ->
            val number = entry.displayName.take(5).toInt()
            require(used.add(number)) { "Duplicate Blu-ray playlist " + number }
            val navigation = try {
                BlurayMplsNavigation.parse(provider.readPlaylist(entry.documentId))
            } catch (ex: IllegalArgumentException) {
                throw IllegalArgumentException("Invalid MPLS playlist " + entry.displayName + ": " + ex.message, ex)
            }
            BluraySafPlaylist(number, entry.documentId, navigation)
        }.sortedBy { it.number }
    }

    fun select(playlists: List<BluraySafPlaylist>, requested: Int? = null): BluraySafPlaylist {
        require(playlists.isNotEmpty()) { "No Blu-ray playlists" }
        if (requested != null) {
            require(requested in 0..99999) { "Invalid Blu-ray playlist number" }
            return playlists.singleOrNull { it.number == requested }
                ?: throw IllegalArgumentException("Blu-ray playlist " + requested + " not found")
        }
        return playlists.sortedWith(
            compareByDescending<BluraySafPlaylist> { it.navigation.durationTicks }
                .thenBy { it.number }
        ).first()
    }
}
