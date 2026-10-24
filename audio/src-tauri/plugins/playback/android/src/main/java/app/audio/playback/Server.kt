// [输入]: Go server 地址、共享访问密钥和设备本地配置。
// [输出]: 原生 HTTP 请求、内存会话与隔离的续播记录。
// [定位]: API 和播放器共用的 server 边界。
// [同步]: ../../../../../../../../folder.md、README.md。
package app.audio.playback

import android.content.Context
import android.content.SharedPreferences
import java.net.CookieManager
import java.net.CookiePolicy
import java.net.HttpURLConnection
import java.net.URI
import java.net.URL
import java.security.MessageDigest
import org.json.JSONObject

internal object Server {
    lateinit var prefs: SharedPreferences
    private val cookies = CookieManager(null, CookiePolicy.ACCEPT_ORIGINAL_SERVER)
    @Volatile var address: String = ""
        private set

    @Synchronized fun init(context: Context) {
        if (::prefs.isInitialized) return
        prefs = context.getSharedPreferences("audio", Context.MODE_PRIVATE)
        address = prefs.getString("server", "")!!
    }

    fun normalize(value: String): String {
        val uri = URI(value.trim())
        require(uri.scheme == "https" || uri.scheme == "http") { "server 地址必须使用 HTTP 或 HTTPS" }
        require(!uri.host.isNullOrEmpty() && uri.rawUserInfo == null && uri.rawQuery == null && uri.rawFragment == null) {
            "server 地址必须包含主机，不能包含凭证、查询参数或片段"
        }
        require(uri.path.isNullOrEmpty() || uri.path == "/") { "server 地址须为根地址，例如 https://audio.example.com" }
        require(uri.port == -1 || uri.port in 1..65535) { "server 端口无效: ${uri.port}" }
        return URI(uri.scheme.lowercase(), null, uri.host.lowercase(), uri.port, null, null, null).toString()
    }

    @Synchronized fun configure(value: String) {
        val next = normalize(value)
        if (next == address) return
        check(prefs.edit().putString("server", next).commit()) { "保存 server 地址失败" }
        cookies.cookieStore.removeAll()
        address = next
    }

    @Synchronized fun cookieHeader(): String = cookies.get(URI(address), emptyMap())["Cookie"]?.joinToString("; ") ?: ""

    // 仅暴露两个既有 API，页面不能借原生桥接请求任意 URL。
    fun request(path: String, key: String?): JSONObject {
        val server = address
        check(server.isNotEmpty()) { "尚未配置 server 地址" }
        require(path == "/api/library" || path == "/api/access") { "不允许请求 path=$path" }
        val connection = URL(server + path).openConnection() as HttpURLConnection
        connection.connectTimeout = 15000
        connection.readTimeout = 20000
        connection.instanceFollowRedirects = false
        connection.setRequestProperty("Cookie", cookieHeader())
        try {
            if (path == "/api/access") {
                require(!key.isNullOrBlank()) { "共享访问密钥不能为空" }
                connection.requestMethod = "POST"
                connection.doOutput = true
                connection.setRequestProperty("Content-Type", "application/json")
                connection.outputStream.use { it.write(JSONObject().put("key", key).toString().toByteArray(Charsets.UTF_8)) }
            }
            val status = connection.responseCode
            synchronized(this) {
                check(server == address) { "请求期间 server 地址已变更: $server" }
                cookies.put(URI(server + path), connection.headerFields.filterKeys { it != null })
            }
            val stream = if (status in 200..299) connection.inputStream else connection.errorStream
            val body = stream?.bufferedReader(Charsets.UTF_8)?.use { it.readText() } ?: ""
            check(server == address) { "请求期间 server 地址已变更: $server" }
            return JSONObject().put("status", status).put("body", body)
        } finally {
            connection.disconnect()
        }
    }

    private fun prefix(): String = MessageDigest.getInstance("SHA-256")
        .digest(address.toByteArray(Charsets.UTF_8)).joinToString("") { "%02x".format(it) } + ":"

    fun position(id: String): Long = prefs.getLong(prefix() + id, 0)
    fun lastId(): String = prefs.getString(prefix() + "last", "")!!

    fun positions(): JSONObject {
        val prefix = prefix()
        val result = JSONObject()
        prefs.all.forEach { (key, value) ->
            val id = key.removePrefix(prefix)
            if (key.startsWith(prefix) && id.matches(Regex("[a-f0-9]{64}")) && value is Long) {
                result.put(id, value / 1000.0)
            }
        }
        return result
    }

    fun save(id: String, position: Long) {
        if (id.isEmpty()) return
        prefs.edit().putLong(prefix() + id, position.coerceAtLeast(0))
            .putString(prefix() + "last", id).apply()
    }
}
