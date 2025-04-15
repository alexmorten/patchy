export function extractMessageId(headers: string): string | null {
  // Message-ID format: <numbers.letters@domain>
  const messageIdRegex = /Message-ID:\s*<([^>]+)>/i;
  const match = headers.match(messageIdRegex);
  
  if (match && match[1]) {
    return match[1];
  }
  
  return null;
}

// Example usage:
// const headers = `... email headers ...`;
// const messageId = extractMessageId(headers);
// console.log(messageId); // "173706974044.1927324.7824600141282028094.stgit@frogsfrogsfrogs" 