/**
 * Validate URL is external
 * @param {string} path - URL path to check
 * @returns {boolean} - True if URL is external
 */
export function isExternal(path: string): boolean {
  if (!path || typeof path !== 'string') return false
  const externalPattern = /^(https?:|mailto:|tel:)/i
  return externalPattern.test(path)
}

/**
 * Validate URL format
 * @param {string} url - URL to validate
 * @returns {boolean} - True if URL is valid
 */
export function isValidURL(url: string): boolean {
  if (!url || typeof url !== 'string') return false

  try {
    const parsedUrl = new URL(url)
    return ['http:', 'https:'].includes(parsedUrl.protocol)
  } catch (e) {
    return false
  }
}

/**
 * Validate username format
 * @param {string} str - Username to validate
 * @returns {boolean} - True if username is valid
 */
export function validateUsername(str: string): boolean {
  const validPattern = /^[a-zA-Z0-9_-]{4,16}$/
  return validPattern.test(str)
}

/**
 * Validate password format
 * @param {string} str - Password to validate
 * @returns {boolean} - True if password is valid
 */
export function validatePassword(str: string): boolean {
  // At least 8 characters, must include letters and numbers
  const validPattern = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/
  return validPattern.test(str)
}

/**
 * Validate email format
 * @param {string} email - Email to validate
 * @returns {boolean} - True if email is valid
 */
export function validateEmail(email: string): boolean {
  if (!email || typeof email !== 'string') return false

  // More comprehensive email validation pattern
  const validPattern = /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
  return validPattern.test(email.toLowerCase())
}