/*
 * WebTransport 客户端实现
 * 用于与 WebTransport 服务器通信
 */

class WebTransportClient {
  constructor(url, onMessageCallback, onErrorCallback) {
    this.url = url;
    this.onMessage = onMessageCallback;
    this.onError = onErrorCallback;
    this.transport = null;
    this.stream = null;
    this.reader = null;
    this.isConnected = false;
  }

  async connect() {
    try {
      // 检查浏览器是否支持 WebTransport
      if (!('WebTransport' in window)) {
        throw new Error('WebTransport is not supported in this browser');
      }

      console.log('Connecting to WebTransport server:', this.url);
      
      // 创建 WebTransport 连接
      this.transport = new WebTransport(this.url);
      
      // 等待连接建立
      await this.transport.ready;
      console.log('WebTransport connection established');

      // 获取双向流
      this.stream = await this.transport.createBidirectionalStream();
      this.reader = this.stream.readable.getReader();
      
      // 开始读取数据
      this.#readLoop();
      
      this.isConnected = true;
      console.log('WebTransport client is ready');
    } catch (error) {
      console.error('WebTransport connection failed:', error);
      this.onError?.(error);
    }
  }

  async #readLoop() {
    try {
      while (true) {
        const { value, done } = await this.reader.read();
        if (done) {
          console.log('WebTransport stream closed');
          break;
        }
        
        // 处理接收到的数据
        if (value) {
          this.onMessage?.(value);
        }
      }
    } catch (error) {
      console.error('Error in WebTransport read loop:', error);
      this.onError?.(error);
    }
  }

  async send(data) {
    if (!this.stream || !this.isConnected) {
      throw new Error('WebTransport is not connected');
    }
    
    const writer = this.stream.writable.getWriter();
    try {
      await writer.write(data);
    } finally {
      writer.releaseLock();
    }
  }

  async disconnect() {
    if (this.reader) {
      await this.reader.cancel();
      this.reader = null;
    }
    
    if (this.transport) {
      await this.transport.close();
      this.transport = null;
    }
    
    this.isConnected = false;
    console.log('WebTransport client disconnected');
  }
}

/*
 * WebTransport 与 WebSocket 混合客户端
 * 自动选择最佳传输协议
 */
class HybridTransportClient {
  constructor(wsUrl, wtUrl, onMessageCallback, onErrorCallback) {
    this.wsUrl = wsUrl;
    this.wtUrl = wtUrl;
    this.onMessage = onMessageCallback;
    this.onError = onErrorCallback;
    this.activeClient = null;
    this.protocol = null; // 'websocket' or 'webtransport'
  }

  async connect() {
    // 优先尝试 WebTransport
    if ('WebTransport' in window) {
      try {
        console.log('Attempting WebTransport connection...');
        this.activeClient = new WebTransportClient(
          this.wtUrl, 
          this.onMessage, 
          (error) => {
            console.warn('WebTransport failed, falling back to WebSocket:', error);
            this.#connectWebSocket();
          }
        );
        await this.activeClient.connect();
        this.protocol = 'webtransport';
        console.log('Connected using WebTransport');
        return;
      } catch (error) {
        console.warn('WebTransport connection failed, falling back to WebSocket:', error);
      }
    }
    
    // 回退到 WebSocket
    this.#connectWebSocket();
  }

  #connectWebSocket() {
    return new Promise((resolve, reject) => {
      try {
        const ws = new WebSocket(this.wsUrl);
        
        ws.binaryType = 'arraybuffer';
        
        ws.onopen = () => {
          console.log('Connected using WebSocket');
          this.activeClient = ws;
          this.protocol = 'websocket';
          resolve();
        };
        
        ws.onmessage = (event) => {
          this.onMessage?.(event.data);
        };
        
        ws.onerror = (error) => {
          this.onError?.(error);
          reject(error);
        };
        
        ws.onclose = () => {
          console.log('WebSocket connection closed');
        };
      } catch (error) {
        this.onError?.(error);
        reject(error);
      }
    });
  }

  async send(data) {
    if (!this.activeClient) {
      throw new Error('No active connection');
    }
    
    if (this.protocol === 'webtransport') {
      await this.activeClient.send(data);
    } else {
      this.activeClient.send(data);
    }
  }

  disconnect() {
    if (this.activeClient) {
      if (this.protocol === 'webtransport') {
        this.activeClient.disconnect();
      } else {
        this.activeClient.close();
      }
    }
  }
}

// 使用示例
/*
const hybridClient = new HybridTransportClient(
  'wss://localhost:8080/ws',  // WebSocket URL
  'https://localhost:4433/wt', // WebTransport URL
  (data) => {
    // 处理接收到的消息
    console.log('Received data:', data);
    
    // 可以解析为视频帧
    try {
      const textDecoder = new TextDecoder();
      const text = textDecoder.decode(data.slice(0, Math.min(100, data.byteLength)));
      console.log('Message preview:', text);
    } catch (e) {
      console.error('Error processing received data:', e);
    }
  },
  (error) => {
    console.error('Connection error:', error);
  }
);

hybridClient.connect()
  .then(() => {
    console.log('Hybrid client connected using:', hybridClient.protocol);
  })
  .catch((error) => {
    console.error('Failed to connect:', error);
  });
*/