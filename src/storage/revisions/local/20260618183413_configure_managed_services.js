const message = 'Configure managed services';

const revision = '20260618183413_configure_managed_services';

async function upgrade() {
  const changes = {};

  changes.enableManagedLocalServices = true;

  changes.storageVersion = revision;
  return browser.storage.local.set(changes);
}

export {message, revision, upgrade};
