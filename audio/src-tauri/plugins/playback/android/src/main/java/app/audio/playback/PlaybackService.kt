// [输入]: 音频队列、原生控制指令和 server 会话。
// [输出]: 后台音频、锁屏控制与设备本地续播记录。
// [定位]: 独立于 WebView 生命周期的播放服务。
// [同步]: ../../../../../../../../folder.md、README.md。
package app.audio.playback

import android.app.PendingIntent
import android.os.Handler
import android.os.Looper
import androidx.media3.common.AudioAttributes
import androidx.media3.common.C
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.common.util.UnstableApi
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.session.MediaSession
import androidx.media3.session.MediaSessionService

@UnstableApi
class PlaybackService : MediaSessionService() {
    private var session: MediaSession? = null
    private val handler = Handler(Looper.getMainLooper())
    private lateinit var player: ExoPlayer
    private var restored = false

    companion object {
        var error: String = ""
        var unauthorized: Boolean = false
        val http = DefaultHttpDataSource.Factory().setConnectTimeoutMs(15000).setReadTimeoutMs(20000)
    }

    private fun save() {
        if (!restored) return
        val item = player.currentMediaItem ?: return
        val ended = player.playbackState == Player.STATE_ENDED
        Server.save(item.mediaId, if (ended) 0 else player.currentPosition)
    }

    private val ticker = object : Runnable {
        override fun run() {
            if (restored) save()
            handler.postDelayed(this, 4000)
        }
    }

    override fun onCreate() {
        super.onCreate()
        Server.init(this)
        player = ExoPlayer.Builder(this)
            .setMediaSourceFactory(DefaultMediaSourceFactory(http))
            .setAudioAttributes(AudioAttributes.Builder().setUsage(C.USAGE_MEDIA)
                .setContentType(C.AUDIO_CONTENT_TYPE_SPEECH).build(), true)
            .setHandleAudioBecomingNoisy(true)
            .setWakeMode(C.WAKE_MODE_NETWORK)
            .build()
        player.addListener(object : Player.Listener {
            override fun onMediaItemTransition(item: MediaItem?, reason: Int) {
                restored = false
                if (item != null) {
                    if (player.playbackState == Player.STATE_READY) restore()
                }
            }

            override fun onPlaybackStateChanged(state: Int) {
                if (state == Player.STATE_READY && !restored) {
                    restore()
                }
                if (state == Player.STATE_ENDED) save()
            }

            override fun onPositionDiscontinuity(old: Player.PositionInfo, next: Player.PositionInfo, reason: Int) {
                val id = old.mediaItem?.mediaId ?: return
                if (reason == Player.DISCONTINUITY_REASON_AUTO_TRANSITION) {
                    Server.save(id, 0)
                } else if (old.mediaItemIndex != next.mediaItemIndex) {
                    Server.save(id, old.positionMs)
                }
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                if (restored) save()
            }

            override fun onPlayerError(failure: PlaybackException) {
                save()
                unauthorized = generateSequence<Throwable>(failure) { it.cause }.any {
                    it is androidx.media3.datasource.HttpDataSource.InvalidResponseCodeException && it.responseCode == 401
                }
                error = if (unauthorized) "访问会话已失效，请重新输入共享访问密钥。"
                    else "播放音频 id=${player.currentMediaItem?.mediaId} 失败: ${failure.errorCodeName}，${failure.cause?.message ?: failure.message}"
                player.pause()
            }
        })
        val builder = MediaSession.Builder(this, player)
        packageManager.getLaunchIntentForPackage(packageName)?.let { intent ->
            builder.setSessionActivity(PendingIntent.getActivity(this, 0, intent,
                PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE))
        }
        session = builder.build()
        handler.postDelayed(ticker, 4000)
    }

    private fun restore() {
        val item = player.currentMediaItem ?: return
        restored = true
        val position = Server.position(item.mediaId)
        if (position > 0 && position < player.duration - 2000) player.seekTo(position)
        Server.save(item.mediaId, player.currentPosition)
    }

    override fun onGetSession(controllerInfo: MediaSession.ControllerInfo): MediaSession? {
        return if (controllerInfo.packageName == packageName || controllerInfo.isTrusted) session else null
    }

    override fun onDestroy() {
        handler.removeCallbacksAndMessages(null)
        save()
        session?.release()
        player.release()
        session = null
        super.onDestroy()
    }
}
