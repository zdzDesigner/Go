// [输入]: 内嵌 H5 的 Tauri 指令。
// [输出]: server API 结果与 Media3 播放控制及状态。
// [定位]: Android 原生桥接，所有网络操作在工作线程执行。
// [同步]: ../../../../../../../../folder.md、README.md。
package app.audio.playback

import android.app.Activity
import android.content.ComponentName
import androidx.core.content.ContextCompat
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.Player
import androidx.media3.common.util.UnstableApi
import androidx.media3.session.MediaController
import androidx.media3.session.SessionToken
import app.tauri.plugin.JSObject
import app.tauri.annotation.Command
import app.tauri.annotation.InvokeArg
import app.tauri.annotation.TauriPlugin
import app.tauri.plugin.Invoke
import app.tauri.plugin.Plugin
import java.util.concurrent.Executors
import org.json.JSONArray

@InvokeArg
class Args {
    var address: String? = null
    var path: String? = null
    var key: String? = null
    var items: String? = null
    var id: String? = null
    var action: String? = null
    var position: Double? = null
}

@UnstableApi
@TauriPlugin
class PlaybackPlugin(private val activity: Activity) : Plugin(activity) {
    private val worker = Executors.newSingleThreadExecutor()
    private var controller: MediaController? = null

    init { Server.init(activity) }

    private fun withPlayer(invoke: Invoke, action: (MediaController) -> Unit) {
        val existing = controller
        if (existing != null && existing.isConnected) {
            try { action(existing) } catch (error: Exception) { invoke.reject("播放操作失败: ${error.message}") }
            return
        }
        val future = MediaController.Builder(activity,
            SessionToken(activity, ComponentName(activity, PlaybackService::class.java))).buildAsync()
        future.addListener({
            try {
                val player = future.get()
                controller = player
                action(player)
            } catch (error: Exception) {
                invoke.reject("连接播放服务失败: ${error.cause?.message ?: error.message}")
            }
        }, ContextCompat.getMainExecutor(activity))
    }

    @Command
    fun configure(invoke: Invoke) {
        try {
            val args = invoke.parseArgs(Args::class.java)
            args.address?.let {
                val normalized = Server.normalize(it)
                if (normalized != Server.address) {
                    controller?.let { player ->
                        player.currentMediaItem?.let { item -> Server.save(item.mediaId, player.currentPosition) }
                        player.stop()
                        player.clearMediaItems()
                    }
                    Server.configure(normalized)
                    PlaybackService.error = ""
                    PlaybackService.unauthorized = false
                }
            }
            invoke.resolve(JSObject().put("address", Server.address))
        } catch (error: Exception) { invoke.reject("配置 server 失败: ${error.message}") }
    }

    @Command
    fun request(invoke: Invoke) {
        val args = invoke.parseArgs(Args::class.java)
        worker.execute {
            try {
                val result = Server.request(requireNotNull(args.path) { "缺少 API path" }, args.key)
                PlaybackService.http.setDefaultRequestProperties(mapOf("Cookie" to Server.cookieHeader()))
                invoke.resolve(JSObject(result.toString()))
            } catch (error: Exception) {
                invoke.reject("请求 server=${Server.address} path=${args.path} 失败: ${error.message}")
            }
        }
    }

    @Command
    fun play(invoke: Invoke) {
        val args = invoke.parseArgs(Args::class.java)
        withPlayer(invoke) { player ->
            check(Server.address.isNotEmpty()) { "尚未配置 server 地址" }
            val items = JSONArray(requireNotNull(args.items) { "缺少播放队列" })
            val queue = (0 until items.length()).map { index ->
                val item = items.getJSONObject(index)
                val id = item.getString("id")
                require(id.matches(Regex("[a-f0-9]{64}"))) { "音频资源 ID 无效: $id" }
                MediaItem.Builder().setMediaId(id).setUri("${Server.address}/api/audio/$id/stream")
                    .setMediaMetadata(MediaMetadata.Builder().setTitle(item.getString("name"))
                        .setArtist(item.getString("path")).build()).build()
            }
            val index = queue.indexOfFirst { it.mediaId == args.id }
            require(index >= 0) { "播放队列中不存在 id=${args.id}" }
            player.currentMediaItem?.let { Server.save(it.mediaId, player.currentPosition) }
            PlaybackService.error = ""
            PlaybackService.unauthorized = false
            PlaybackService.http.setDefaultRequestProperties(mapOf("Cookie" to Server.cookieHeader()))
            player.setMediaItems(queue, index, 0)
            player.prepare()
            player.play()
            invoke.resolve()
        }
    }

    @Command
    fun control(invoke: Invoke) {
        val args = invoke.parseArgs(Args::class.java)
        withPlayer(invoke) { player ->
            when (args.action) {
                "pause" -> player.pause()
                "resume" -> {
                    if (player.playbackState == Player.STATE_IDLE) player.prepare()
                    player.play()
                }
                "seek" -> {
                    val position = requireNotNull(args.position) { "缺少播放位置" }
                    require(position.isFinite() && position >= 0) { "播放位置无效: $position" }
                    player.seekTo((position * 1000).toLong())
                    player.currentMediaItem?.let { Server.save(it.mediaId, player.currentPosition) }
                }
                "previous" -> player.seekToPreviousMediaItem()
                "next" -> player.seekToNextMediaItem()
                else -> throw IllegalArgumentException("未知播放操作 action=${args.action}")
            }
            invoke.resolve()
        }
    }

    @Command
    fun state(invoke: Invoke) {
        withPlayer(invoke) { player ->
            val item = player.currentMediaItem
            val positions = Server.positions()
            invoke.resolve(JSObject().put("id", item?.mediaId ?: Server.lastId())
                .put("active", item != null).put("playing", player.playWhenReady)
                .put("position", if (item != null) player.currentPosition / 1000.0 else Server.position(Server.lastId()) / 1000.0)
                .put("duration", player.duration.coerceAtLeast(0) / 1000.0)
                .put("previous", player.hasPreviousMediaItem()).put("next", player.hasNextMediaItem())
                .put("positions", positions).put("error", PlaybackService.error)
                .put("unauthorized", PlaybackService.unauthorized))
        }
    }

    override fun onDestroy(activity: androidx.appcompat.app.AppCompatActivity) {
        controller?.release()
        controller = null
        worker.shutdown()
    }
}
