/** Client-only proof helpers — images are never uploaded to object storage. */

export const MAX_PROOF_BYTES = 3 * 1024 * 1024 // 3MB

export function isImageFile(file: File): boolean {
  return file.type.startsWith('image/')
}

/** Read file as data URL in the browser (for AI vision only; do not POST to join as proof body). */
export function fileToDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    if (file.size > MAX_PROOF_BYTES) {
      reject(new Error('Image must be under 3 MB'))
      return
    }
    if (!isImageFile(file)) {
      reject(new Error('Please choose an image (JPG, PNG, WebP)'))
      return
    }
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(new Error('Unable to read image'))
    reader.readAsDataURL(file)
  })
}

/** Lightweight marker saved on participation instead of the image. */
export function proofMarker(fileName: string): string {
  const safe = fileName.replace(/[^\w.\-]+/g, '_').slice(0, 120)
  return `upload:${safe || 'proof.jpg'}`
}
