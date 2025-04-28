/**
 * Validate URL is external
 * @param {string} path - URL path to check
 * @returns {boolean} - True if URL is external
 */
export function isExternal(path: string): boolean {
  const externalPattern = /^(https?:|mailto:|tel:)/
  return externalPattern.test(path)
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
  const validPattern = /^[a-zA-Z0-9_-]+@[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+)+$/
  return validPattern.test(email)
}