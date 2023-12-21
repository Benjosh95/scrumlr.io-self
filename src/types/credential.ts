export interface Credential {
    ID: string;
    publicKey: string;
    attestationType: string;
    transport: string[]; // Use 'usb' | 'nfc' | 'ble' | 'internal' | 'platform' as appropriate
    flags: CredentialFlags;
    authenticator: Authenticator;
  }
  
  export interface CredentialFlags {
    userPresent: boolean;
    userVerified: boolean;
    backupEligible: boolean;
    backupState: boolean;
  }
  
  export interface Authenticator {
    aaguid: string;
    signCount: number;
    attachment: string; // Use 'platform' | 'cross-platform' as appropriate
    cloneWarning: string;
  }

  
  