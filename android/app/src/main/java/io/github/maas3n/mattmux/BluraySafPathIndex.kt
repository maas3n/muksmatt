package io.github.maas3n.mattmux

import java.util.Locale

/**
 * Pure path and directory-index rules shared by SAF and the libbluray
 * bd_open_files callback bridge. No real filesystem paths are exposed.
 */
internal class BluraySafPathIndex {
    internal data class Node(
        val documentId: String,
        val name: String,
        val isDirectory: Boolean,
    )

    private val nodes = LinkedHashMap<String, Node>()
    private val children = HashMap<String, MutableList<String>>()

    companion object {
        private val segment = Regex("[A-Za-z0-9_.-]{1,200}")

        fun canonical(input: String): String? {
            if (input.length > 1024 || '\u0000' in input || '\\' in input) return null
            val path = input.trimStart('/')
            if (path.isEmpty()) return ""
            if (path.endsWith('/')) return null
            val parts = path.split('/')
            if (parts.size > 20 || parts.any {
                    it == "." || it == ".." || !segment.matches(it)
                }) return null
            return parts.joinToString("/") { it.uppercase(Locale.ROOT) }
        }
    }

    fun register(path: String, documentId: String, name: String, directory: Boolean) {
        val key = canonical(path) ?: throw IllegalArgumentException("Invalid Blu-ray path")
        require(documentId.isNotBlank() && nodes.size < 30000) { "Too many Blu-ray documents" }
        require(key !in nodes) { "Ambiguous Blu-ray filename: $path" }
        if (key.isNotEmpty()) {
            val parent = key.substringBeforeLast('/', "")
            require(nodes[parent]?.isDirectory == true) { "Missing Blu-ray directory: $parent" }
            require(segment.matches(name) && name != "." && name != "..") {
                "Unsafe Blu-ray document name"
            }
            children.getOrPut(parent) { mutableListOf() }.add(name)
        }
        nodes[key] = Node(documentId, name, directory)
    }

    fun node(path: String): Node? = canonical(path)?.let { nodes[it] }

    fun names(path: String): Array<String>? {
        val key = canonical(path) ?: return null
        if (nodes[key]?.isDirectory != true) return null
        return children[key]?.toTypedArray() ?: emptyArray()
    }
}
