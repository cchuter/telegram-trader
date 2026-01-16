import SignClient from '@walletconnect/sign-client';
import { SessionTypes } from '@walletconnect/types';
import * as QRCode from 'qrcode-terminal';

/**
 * WalletConnect manager for handling wallet connections
 * Supports both WalletConnect v2 and manual key input
 */
export class WalletConnectManager {
  private signClient: SignClient | null = null;
  private sessions: Map<string, SessionTypes.Struct> = new Map();
  private projectId: string;

  constructor(projectId: string) {
    this.projectId = projectId;
  }

  /**
   * Initialize WalletConnect SignClient
   */
  async initialize(): Promise<void> {
    if (this.signClient) {
      return;
    }

    this.signClient = await SignClient.init({
      projectId: this.projectId,
      metadata: {
        name: 'Telegram Trader',
        description: 'Telegram bot for GalaChain trading',
        url: 'https://telegram-trader.app',
        icons: ['https://telegram-trader.app/icon.png'],
      },
    });

    // Set up event listeners
    this.setupEventListeners();
  }

  /**
   * Set up WalletConnect event listeners
   */
  private setupEventListeners(): void {
    if (!this.signClient) {
      throw new Error('SignClient not initialized');
    }

    // Listen for session proposals
    this.signClient.on('session_proposal', async (proposal) => {
      console.log('Session proposal received:', proposal);
    });

    // Listen for session updates
    this.signClient.on('session_update', ({ topic, params }) => {
      console.log('Session updated:', topic, params);
      const session = this.signClient?.session.get(topic);
      if (session) {
        this.sessions.set(topic, session);
      }
    });

    // Listen for session deletions
    this.signClient.on('session_delete', ({ topic }) => {
      console.log('Session deleted:', topic);
      this.sessions.delete(topic);
    });

    // Listen for session events
    this.signClient.on('session_event', ({ topic, params }) => {
      console.log('Session event:', topic, params);
    });
  }

  /**
   * Create a new WalletConnect session
   * Returns the URI for QR code and session ID
   */
  async createSession(userId: number): Promise<{
    uri: string;
    sessionId: string;
    deepLink: string;
  }> {
    if (!this.signClient) {
      await this.initialize();
    }

    if (!this.signClient) {
      throw new Error('Failed to initialize SignClient');
    }

    try {
      // Create session proposal
      const { uri, approval } = await this.signClient.connect({
        requiredNamespaces: {
          eip155: {
            methods: [
              'eth_sendTransaction',
              'eth_signTransaction',
              'eth_sign',
              'personal_sign',
              'eth_signTypedData',
            ],
            chains: ['eip155:1'], // Ethereum mainnet (GalaChain uses similar structure)
            events: ['chainChanged', 'accountsChanged'],
          },
        },
      });

      if (!uri) {
        throw new Error('Failed to generate WalletConnect URI');
      }

      // Generate QR code in terminal for debugging
      console.log('WalletConnect URI generated for user', userId);
      QRCode.generate(uri, { small: true });

      // Wait for session approval in background (don't block)
      approval()
        .then((session) => {
          console.log('Session approved:', session.topic);
          this.sessions.set(session.topic, session);
          // TODO: Save session to database with userId
        })
        .catch((error) => {
          console.error('Session approval failed:', error);
        });

      // Generate deep link for mobile wallets
      const deepLink = `gala://wc?uri=${encodeURIComponent(uri)}`;

      return {
        uri,
        sessionId: uri.split('@')[0].replace('wc:', ''), // Extract session ID from URI
        deepLink,
      };
    } catch (error) {
      console.error('Error creating WalletConnect session:', error);
      throw error;
    }
  }

  /**
   * Get an active session by session ID
   */
  getSession(sessionId: string): SessionTypes.Struct | undefined {
    return this.sessions.get(sessionId);
  }

  /**
   * Disconnect a session
   */
  async disconnectSession(sessionId: string): Promise<void> {
    if (!this.signClient) {
      throw new Error('SignClient not initialized');
    }

    const session = this.sessions.get(sessionId);
    if (session) {
      await this.signClient.disconnect({
        topic: session.topic,
        reason: {
          code: 6000,
          message: 'User disconnected',
        },
      });
      this.sessions.delete(sessionId);
    }
  }

  /**
   * Get wallet address from session
   */
  getWalletAddress(sessionId: string): string | null {
    const session = this.sessions.get(sessionId);
    if (!session) {
      return null;
    }

    // Extract address from session accounts
    const accounts = session.namespaces.eip155?.accounts || [];
    if (accounts.length > 0) {
      // Account format: "eip155:1:0x..."
      const address = accounts[0].split(':')[2];
      return address;
    }

    return null;
  }

  /**
   * Check if a session is active
   */
  isSessionActive(sessionId: string): boolean {
    return this.sessions.has(sessionId);
  }
}

/**
 * Manual wallet storage for private key fallback
 * In production, this should use encrypted storage
 */
export interface ManualWallet {
  userId: number;
  address: string;
  publicKey: string;
  privateKey: string; // Should be encrypted in production!
  createdAt: Date;
}

/**
 * Manager for manual wallet connections (private key input)
 */
export class ManualWalletManager {
  private wallets: Map<number, ManualWallet> = new Map();

  /**
   * Store a manual wallet connection
   * WARNING: Private keys should be encrypted in production!
   */
  storeWallet(
    userId: number,
    privateKey: string,
    publicKey: string,
    address: string
  ): ManualWallet {
    const wallet: ManualWallet = {
      userId,
      address,
      publicKey,
      privateKey, // In production: encrypt this!
      createdAt: new Date(),
    };

    this.wallets.set(userId, wallet);
    console.log(`Manual wallet stored for user ${userId}`);

    // TODO: Save to database with encryption
    return wallet;
  }

  /**
   * Get wallet for a user
   */
  getWallet(userId: number): ManualWallet | undefined {
    return this.wallets.get(userId);
  }

  /**
   * Remove wallet for a user
   */
  removeWallet(userId: number): boolean {
    return this.wallets.delete(userId);
  }

  /**
   * Check if user has a wallet
   */
  hasWallet(userId: number): boolean {
    return this.wallets.has(userId);
  }
}
