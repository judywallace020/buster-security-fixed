import {
  pipeline,
  env as transformersEnv,
  LogLevel,
  ModelRegistry
} from '@huggingface/transformers';

import {noPunctuationRx} from 'utils/data';

function getPromptApiParams() {
  return {
    expectedInputs: [{type: 'text', languages: ['en']}, {type: 'audio'}],
    expectedOutputs: [{type: 'text', languages: ['en']}]
  };
}

async function initPromptApi({params = null} = {}) {
  if (!params) {
    params = getPromptApiParams();
  }

  try {
    const session = await LanguageModel.create(params);
    await session.destroy();
  } catch (err) {
    console.log(err.toString());
  }
}

async function getPromptApiStatus({params = null, init = false} = {}) {
  if (!params) {
    params = getPromptApiParams();
  }

  let status;

  try {
    if (typeof LanguageModel !== 'undefined') {
      status = await LanguageModel.availability(params);

      if (
        init &&
        !['unavailable', 'available', 'downloading'].includes(status)
      ) {
        initPromptApi({params});
      }
    }
  } catch (err) {
    console.log(err.toString());
  }

  return status ? status : 'unavailable';
}

async function getPromptApiSession({params = null} = {}) {
  if (!params) {
    params = getPromptApiParams();
  }

  try {
    return await LanguageModel.create(params);
  } catch (err) {
    console.log(err.toString());
  }
}

async function getPromptApiResult(audioContent) {
  const result = {};

  const session = await getPromptApiSession();

  if (session) {
    try {
      const data = await session.prompt([
        {
          role: 'user',
          content: [
            {
              type: 'text',
              value:
                'transcribe this audio, the output cannot contain punctuation or uppercase letters'
            },
            {type: 'audio', value: audioContent}
          ]
        }
      ]);

      if (data) {
        result.text = data.trim();
      }
    } catch (err) {
      console.log(err.toString());
    } finally {
      session.destroy();
    }
  }

  return result;
}

function getTransformersApiParams({callback = null} = {}) {
  const options = {device: 'wasm', dtype: 'q8'};

  if (callback) {
    options.progress_callback = callback;
  }

  return ['automatic-speech-recognition', 'Xenova/whisper-tiny', options];
}

function setTransformersApiEnv() {
  transformersEnv.logLevel = LogLevel.ERROR;
  transformersEnv.backends.onnx.wasm.wasmPaths =
    browser.runtime.getURL('src/wasm/');
}

async function initTransformersApi({params = null, callback = null} = {}) {
  setTransformersApiEnv();

  if (!params) {
    params = getTransformersApiParams({callback});
  }

  try {
    const session = await pipeline(...params);
    await session.dispose();
  } catch (err) {
    console.log(err.toString());
  }
}

async function getTransformersApiStatus({params = null, init = false} = {}) {
  setTransformersApiEnv();

  if (!params) {
    params = getTransformersApiParams();
  }

  let isCached;

  try {
    isCached = await ModelRegistry.is_pipeline_cached(...params);

    if (init && !isCached) {
      initTransformersApi({params});
    }
  } catch (err) {
    console.log(err.toString());
  }

  return isCached ? 'available' : 'downloadable';
}

async function getTransformersApiSession({params = null} = {}) {
  setTransformersApiEnv();

  if (!params) {
    params = getTransformersApiParams();
  }

  try {
    return await pipeline(...params);
  } catch (err) {
    console.log(err.toString());
  }
}

async function getManagedLocalServiceStatus() {
  let status = await getPromptApiStatus();

  const data = {};

  if (status !== 'unavailable') {
    data.modelId = 'geminiNano';
  } else {
    data.modelId = 'whisper';
    status = await getTransformersApiStatus();
  }

  if (status === 'available') {
    status = 'ready to use';
  } else if (status === 'downloadable') {
    status = 'ready to download';
  }

  data.status = status;

  return data;
}

async function getTransformersApiResult(audioContent) {
  const result = {};

  const session = await getTransformersApiSession();

  if (session) {
    try {
      const data = await session(audioContent);

      if (data) {
        result.text = data.text
          .replace(noPunctuationRx, '')
          .toLowerCase()
          .trim();
      }
    } catch (err) {
      console.log(err.toString());
    } finally {
      session.dispose();
    }
  }

  return result;
}

export {
  getPromptApiParams,
  initPromptApi,
  getPromptApiStatus,
  getPromptApiSession,
  getPromptApiResult,
  getTransformersApiParams,
  setTransformersApiEnv,
  initTransformersApi,
  getTransformersApiStatus,
  getTransformersApiSession,
  getManagedLocalServiceStatus,
  getTransformersApiResult
};
