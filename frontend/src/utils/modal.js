// Form handlers close their own dialog only after a successful submission.
// Returning false prevents Arco from discarding input on validation/network errors.
export async function keepModalOpen(submit) {
  try {
    await submit();
  } catch {
    /* Request layer already displays the API error. */
  }
  return false;
}
