<template>
  <vn-app v-if="dataLoaded" :class="appClasses">
    <div class="section-services">
      <div class="section-title" v-once>
        {{ getText('optionSectionTitle_services') }}
      </div>
      <div class="option-wrap">
        <div class="option select">
          <vn-select
            :label="getText('optionTitle_speechService')"
            :items="listItems.speechService"
            v-model="options.speechService"
            transition="scale-transition"
          >
          </vn-select>
        </div>

        <div class="option" v-if="options.speechService === 'managed'">
          <vn-switch
            :label="getText('optionTitle_localServices')"
            v-model="options.enableManagedLocalServices"
          ></vn-switch>
        </div>

        <div
          class="model-card"
          v-if="
            options.speechService === 'managed' &&
            options.enableManagedLocalServices &&
            localServiceStatus
          "
        >
          <div class="model-desc">
            <div>
              Model: {{ getText(`modelName_${localServiceStatus.modelId}`) }}
            </div>

            <div>Status: {{ localServiceStatus.status }}</div>
          </div>

          <vn-button
            class="download-button"
            v-if="
              localServiceStatus.modelId === 'whisper' &&
              localServiceStatus.status !== 'ready to use'
            "
            :disabled="localServiceStatus.status !== 'ready to download'"
            @click="downloadTransformersApiModel"
            variant="elevated"
          >
            {{ getText('buttonLabel_downloadModel') }}
          </vn-button>
        </div>

        <div class="option" v-if="options.speechService === 'managed'">
          <vn-switch
            :label="getText('optionTitle_remoteServices')"
            v-model="options.enableManagedRemoteServices"
          ></vn-switch>
        </div>

        <a
          class="service-guide"
          v-if="options.speechService === 'managed'"
          target="_blank"
          rel="noreferrer"
          href="https://github.com/dessant/buster/wiki/Managed-services"
        >
          {{ getText('linkText_learnMore') }}
        </a>

        <div
          class="option text-field"
          v-if="options.speechService === 'googleSpeechApi'"
        >
          <vn-text-field
            :label="getText('inputLabel_apiKey')"
            v-model.trim="options.googleSpeechApiKey"
          >
          </vn-text-field>
        </div>

        <a
          class="service-guide"
          v-if="options.speechService === 'googleSpeechApi'"
          target="_blank"
          rel="noreferrer"
          href="https://github.com/dessant/buster/wiki/Configuring-Google-Cloud-Speech-to-Text"
        >
          {{ getText('linkText_apiGuide') }}
        </a>

        <div
          class="option text-field"
          v-if="options.speechService === 'ibmSpeechApi'"
        >
          <vn-text-field
            v-model.trim="options.ibmSpeechApiUrl"
            :label="getText('inputLabel_apiUrl')"
          >
          </vn-text-field>
        </div>
        <div
          class="option text-field"
          v-if="options.speechService === 'ibmSpeechApi'"
        >
          <vn-text-field
            v-model.trim="options.ibmSpeechApiKey"
            :label="getText('inputLabel_apiKey')"
          >
          </vn-text-field>
        </div>

        <a
          class="service-guide"
          v-if="options.speechService === 'ibmSpeechApi'"
          target="_blank"
          rel="noreferrer"
          href="https://github.com/dessant/buster/wiki/Configuring-IBM-Watson-Speech-to-Text"
        >
          {{ getText('linkText_apiGuide') }}
        </a>

        <div
          class="option select"
          v-if="options.speechService === 'microsoftSpeechApi'"
        >
          <vn-select
            :label="getText('optionTitle_microsoftSpeechApiLoc')"
            :items="listItems.microsoftSpeechApiLoc"
            v-model="options.microsoftSpeechApiLoc"
            transition="scale-transition"
          >
          </vn-select>
        </div>
        <div
          class="option text-field"
          v-if="options.speechService === 'microsoftSpeechApi'"
        >
          <vn-text-field
            v-model.trim="options.microsoftSpeechApiKey"
            :label="getText('inputLabel_apiKey')"
          >
          </vn-text-field>
        </div>

        <a
          class="service-guide"
          v-if="options.speechService === 'microsoftSpeechApi'"
          target="_blank"
          rel="noreferrer"
          href="https://github.com/dessant/buster/wiki/Configuring-Microsoft-Azure-Speech-to-Text"
        >
          {{ getText('linkText_apiGuide') }}
        </a>

        <vn-text-field
          class="text-field"
          v-if="options.speechService === 'witSpeechApi'"
          v-for="item in witSpeechApis"
          :key="item.id"
          :model-value="options.witSpeechApiKeys[item] || ''"
          :label="
            getText('inputLabel_apiKeyType', [
              getText(`optionValue_witSpeechApiLang_${item}`)
            ])
          "
          @update:modelValue="saveWitSpeechApiKey($event.trim(), item)"
        >
        </vn-text-field>
        <div
          class="wit-add-api"
          v-if="options.speechService === 'witSpeechApi'"
        >
          <vn-select
            :label="getText('optionTitle_witSpeechApiLang')"
            :items="listItems.witSpeechApiLang"
            v-model="witSpeechApiLang"
            transition="scale-transition"
          >
          </vn-select>

          <vn-button :disabled="!witSpeechApiLang" @click="addWitSpeechApi">
            {{ getText('buttonLabel_addApi') }}
          </vn-button>
        </div>

        <a
          class="service-guide"
          v-if="options.speechService === 'witSpeechApi'"
          target="_blank"
          rel="noreferrer"
          href="https://github.com/dessant/buster/wiki/Configuring-Wit.ai"
        >
          {{ getText('linkText_apiGuide') }}
        </a>
      </div>
    </div>

    <div class="section-client">
      <div class="section-title" v-once>
        {{ getText('optionSectionTitle_client') }}
      </div>
      <div class="section-desc" v-once>
        {{ getText('optionSectionDescription_client') }}
      </div>
      <div class="option-wrap">
        <div
          class="option"
          v-if="
            clientAppInstalled ||
            (clientAppVerified && options.simulateUserInput)
          "
        >
          <vn-switch
            :label="getText('optionTitle_simulateUserInput')"
            v-model="options.simulateUserInput"
          ></vn-switch>
        </div>

        <div
          class="option"
          v-if="clientAppVerified && options.simulateUserInput"
        >
          <vn-switch
            :label="getText('optionTitle_navigateWithKeyboard')"
            v-model="options.navigateWithKeyboard"
          ></vn-switch>
        </div>

        <div class="option" v-if="clientAppInstalled">
          <vn-switch
            :label="getText('optionTitle_autoUpdateClientApp')"
            v-model="options.autoUpdateClientApp"
          ></vn-switch>
        </div>

        <div
          class="client-download"
          v-if="clientAppVerified && !clientAppInstalled"
        >
          <div
            class="download-desc"
            v-html="
              getText('pageContent_optionClientAppDownloadDesc', [
                `<a target='_blank' rel='noreferrer' href='${installGuideUrl}'>${getText(
                  'linkText_installGuide'
                )}</a>`
              ])
            "
          ></div>
          <div class="download-error" v-if="!clientAppDownloadUrl">
            {{ getText('pageContent_optionClientAppOSError') }}
          </div>

          <vn-button
            class="download-button"
            :disabled="!clientAppDownloadUrl"
            @click="$refs.dlLink.click()"
            variant="elevated"
          >
            {{ getText('buttonLabel_downloadApp') }}
          </vn-button>
          <a
            ref="dlLink"
            class="download-link"
            target="_blank"
            rel="noreferrer"
            :href="clientAppDownloadUrl"
          ></a>
        </div>
      </div>
    </div>

    <div class="section-misc">
      <div class="section-title" v-once>
        {{ getText('optionSectionTitle_misc') }}
      </div>
      <div class="option-wrap">
        <div class="option select">
          <vn-select
            :label="getText('optionTitle_appTheme')"
            :items="listItems.appTheme"
            v-model="options.appTheme"
            transition="scale-transition"
          >
          </vn-select>
        </div>
        <div class="option">
          <vn-switch
            :label="getText('optionTitle_loadEnglishChallenge')"
            v-model="options.loadEnglishChallenge"
          ></vn-switch>
        </div>
        <div class="option" v-if="!options.loadEnglishChallenge">
          <vn-switch
            :label="getText('optionTitle_tryEnglishSpeechModel')"
            v-model="options.tryEnglishSpeechModel"
          ></vn-switch>
        </div>
        <div class="option" v-if="contributionsEnabled">
          <vn-switch
            :label="getText('optionTitle_showContribPage')"
            v-model="options.showContribPage"
          ></vn-switch>
        </div>
      </div>
    </div>

    <div
      class="section-sponsors"
      v-if="contributionsEnabled || sponsorsEnabled"
    >
      <div class="section-title" v-once>
        {{ getText('optionSectionTitle_sponsors') }}
      </div>
      <div class="option-wrap">
        <div
          class="option sponsor-logo"
          v-if="sponsorsEnabled"
          v-for="(item, index) in sponsors"
          :key="index"
        >
          <a
            :href="getSponsorUrl(item)"
            @click.prevent="showSponsor(item)"
            @keyup.enter.prevent="showSponsor(item)"
          >
            <img :src="getSponsorLogo(item, {variant: theme})" />
          </a>
        </div>
        <div class="option button" v-if="contributionsEnabled">
          <vn-button
            class="contribute-button vn-icon--start"
            @click="showContribute"
            ><vn-icon
              src="/src/assets/icons/misc/favorite-filled.svg"
            ></vn-icon>
            {{ getText('buttonLabel_contribute') }}
          </vn-button>
        </div>
      </div>
    </div>
  </vn-app>
</template>

<script>
import {toRaw} from 'vue';
import {App, Button, Icon, Select, Switch, TextField} from 'vueton';

import storage from 'storage/storage';
import {
  getListItems,
  showContributePage,
  pingClientApp,
  showSponsorPage,
  getAppTheme,
  getSponsorUrl,
  getSponsorLogo
} from 'utils/app';
import {initTransformersApi, getManagedLocalServiceStatus} from 'utils/models';
import {getText} from 'utils/common';
import {
  enableContributions,
  enableSponsors,
  clientAppVersion
} from 'utils/config';
import {
  optionKeys,
  clientAppPlatforms,
  captchaWitSpeechApiLangCodes,
  microsoftSpeechApiRegions,
  sponsors
} from 'utils/data';

export default {
  components: {
    [App.name]: App,
    [Button.name]: Button,
    [Icon.name]: Icon,
    [Select.name]: Select,
    [Switch.name]: Switch,
    [TextField.name]: TextField
  },

  data: function () {
    return {
      dataLoaded: false,

      listItems: {
        ...getListItems(
          {
            speechService: [
              'managed',
              'googleSpeechApi',
              'witSpeechApi',
              'ibmSpeechApi',
              'microsoftSpeechApi'
            ]
          },
          {scope: 'optionValue_speechService'}
        ),
        ...getListItems(
          {microsoftSpeechApiLoc: microsoftSpeechApiRegions},
          {scope: 'optionValue_microsoftSpeechApiLoc'}
        ),
        ...getListItems(
          {
            witSpeechApiLang: [
              ...new Set(
                Object.values(captchaWitSpeechApiLangCodes).filter(Boolean)
              )
            ].sort()
          },
          {scope: 'optionValue_witSpeechApiLang'}
        ),
        ...getListItems(
          {appTheme: ['auto', 'light', 'dark']},
          {scope: 'optionValue_appTheme'}
        )
      },

      sponsors,

      witSpeechApiLang: null,
      witSpeechApis: [],

      clientAppVerified: false,
      clientAppInstalled: false,
      clientAppDownloadUrl: '',
      installGuideUrl: '',
      contributionsEnabled: true,
      sponsorsEnabled: true,
      localServiceStatus: null,

      theme: '',

      options: {
        speechService: '',
        googleSpeechApiKey: '',
        ibmSpeechApiUrl: '',
        ibmSpeechApiKey: '',
        microsoftSpeechApiLoc: '',
        microsoftSpeechApiKey: '',
        witSpeechApiKeys: {},
        loadEnglishChallenge: false,
        tryEnglishSpeechModel: false,
        simulateUserInput: false,
        autoUpdateClientApp: false,
        navigateWithKeyboard: false,
        appTheme: '',
        showContribPage: false,
        enableManagedLocalServices: false,
        enableManagedRemoteServices: false
      }
    };
  },

  computed: {
    appClasses: function () {
      return {
        'show-sponsors': this.sponsorsEnabled || this.contributionsEnabled
      };
    }
  },

  methods: {
    getText,

    getSponsorUrl,
    getSponsorLogo,

    setup: async function () {
      browser.runtime.onMessage.addListener(this.onMessage);

      const options = await storage.get(optionKeys);

      for (const option of Object.keys(this.options)) {
        this.options[option] = options[option];

        this.$watch(
          `options.${option}`,
          async function (value) {
            if (option.startsWith('enableManaged')) {
              if (
                this.options.enableManagedLocalServices ||
                this.options.enableManagedRemoteServices
              ) {
                await storage.set({
                  enableManagedLocalServices: toRaw(
                    this.options.enableManagedLocalServices
                  ),
                  enableManagedRemoteServices: toRaw(
                    this.options.enableManagedRemoteServices
                  )
                });
                await browser.runtime.sendMessage({id: 'optionChange'});
              }
            } else {
              await storage.set({[option]: toRaw(value)});
              await browser.runtime.sendMessage({id: 'optionChange'});
            }
          },
          {deep: true}
        );
      }

      this.witSpeechApis = Object.keys(options.witSpeechApiKeys);
      this.setWitSpeechApiLangOptions();

      document.title = getText('pageTitle', [
        getText('pageTitle_options'),
        getText('extensionName')
      ]);

      this.sponsorsEnabled = enableSponsors && !!this.sponsors.length;
      this.contributionsEnabled = enableContributions;

      this.theme = await getAppTheme(options.appTheme);
      document.addEventListener('themeChange', ev => {
        this.theme = ev.detail;
      });

      this.verifyClientApp();
      this.verifyLocalServices();

      this.dataLoaded = true;
    },

    onMessage: function (request, sender) {
      if (request.id === 'reloadOptionsPage') {
        self.location.reload();
      }
    },

    verifyClientApp: async function () {
      try {
        await pingClientApp();
        this.clientAppInstalled = true;
      } catch (err) {
        if (!this.installGuideUrl) {
          this.installGuideUrl =
            'https://github.com/dessant/buster/wiki/Installing-the-client-app';
          let {os, arch} = this.$env;
          if (clientAppPlatforms.includes(`${os}/${arch}`)) {
            this.installGuideUrl += `#${os}`;

            if (os === 'macos' && arch === 'arm') {
              arch = 'amd64';
            }
            this.clientAppDownloadUrl = `https://github.com/dessant/buster-client/releases/download/v${clientAppVersion}/buster-client-setup-v${clientAppVersion}-${os}-${arch}`;
            if (os === 'windows') {
              this.clientAppDownloadUrl += '.exe';
            }
          }
        }

        this.clientAppInstalled = false;
      }

      this.clientAppVerified = true;
    },

    verifyLocalServices: async function () {
      this.localServiceStatus = await getManagedLocalServiceStatus();
    },

    setWitSpeechApiLangOptions: function () {
      this.listItems.witSpeechApiLang = this.listItems.witSpeechApiLang.filter(
        item => !this.witSpeechApis.includes(item.value)
      );
    },

    addWitSpeechApi: function () {
      this.witSpeechApis.push(this.witSpeechApiLang);
      this.witSpeechApiLang = null;
      this.setWitSpeechApiLangOptions();
    },

    saveWitSpeechApiKey: function (value, lang) {
      const apiKeys = this.options.witSpeechApiKeys;
      if (value) {
        this.options.witSpeechApiKeys = Object.assign({}, apiKeys, {
          [lang]: value
        });
      } else if (apiKeys[lang]) {
        delete apiKeys[lang];
        this.options.witSpeechApiKeys = Object.assign({}, apiKeys);
      }
    },

    downloadTransformersApiModel: async function () {
      await initTransformersApi({callback: this.modelDownloadProgress});
      this.verifyLocalServices();
    },

    modelDownloadProgress: function (status) {
      if (status.status === 'progress_total') {
        this.localServiceStatus.status = `${Math.trunc(status.progress)}%`;
      }
    },

    showContribute: async function () {
      await showContributePage();
    },

    showSponsor: async function (name) {
      await showSponsorPage({name});
    }
  },

  created: function () {
    document.title = getText('pageTitle', [
      getText('pageTitle_options'),
      getText('extensionName')
    ]);

    this.setup();
  }
};
</script>

<style lang="scss">
@use 'vueton/styles' as vueton;

@include vueton.theme-base;
@include vueton.transitions;

.v-application__wrap {
  display: grid;
  grid-row-gap: 32px;
  grid-column-gap: 48px;
  padding: 24px;
  grid-auto-rows: min-content;
  grid-auto-columns: min-content;
}

.section-title {
  font-size: 20px;
  font-weight: 500;
  letter-spacing: 0.25px;
  line-height: 32px;
}

.section-desc {
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0.25px;
  line-height: 20px;

  padding-top: 8px;
  max-width: 380px;
}

.option-wrap {
  display: grid;
  grid-row-gap: 24px;
  padding-top: 24px;
}

.option {
  display: flex;
  align-items: center;
  height: 20px;

  &.button {
    height: 40px;
  }

  &.select,
  &.text-field {
    height: 56px;
  }
}

.text-field .v-input__control {
  width: 326px;
}

.section-client .section-desc {
  width: 272px;
}

.wit-add-api {
  display: flex;
  align-items: center;

  & .vn-select {
    & .v-input__control,
    & .v-input__details {
      max-width: calc(100vw - 48px - 124px) !important;
    }
  }

  & .vn-button {
    margin-left: 24px;
    @include vueton.theme-prop(color, primary);
  }
}

.service-guide {
  font-size: 16px;
  font-weight: 400;
  letter-spacing: 0.5px;
  line-height: 24px;
}

.client-download,
.model-card {
  width: 272px;
}

.model-card {
  & .download-button {
    margin-bottom: 8px;
  }
}

.model-desc {
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0.25px;
  line-height: 24px;
}

.download-desc,
.download-error {
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0.25px;
  line-height: 20px;
}

.download-desc a {
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0.25px;
  line-height: 20px;
}

.download-desc {
  max-width: 240px;
}

.download-error {
  margin-top: 12px;
  @include vueton.theme-prop(color, error);
}

.download-link {
  visibility: hidden;
}

.download-button {
  margin-top: 24px;
  @include vueton.theme-prop(background-color, primary);

  & .v-btn__content {
    @include vueton.theme-prop(color, on-primary);
  }
}

.section-sponsors {
  & .sponsor-logo,
  & .sponsor-logo a,
  & .sponsor-logo img {
    height: 42px;
  }

  & .sponsor-logo img {
    cursor: pointer;
  }

  & .contribute-button {
    @include vueton.theme-prop(color, primary);

    & .vn-icon {
      @include vueton.theme-prop(background-color, cta);
    }
  }

  & .button:not(:only-child) {
    margin-top: 12px;
  }
}

@media (min-width: 736px) {
  .show-sponsors {
    & .v-application__wrap {
      grid-template-columns: minmax(280px, max-content) max-content;
      grid-template-rows: min-content min-content 1fr;
      grid-template-areas:
        'services sponsors'
        'client sponsors'
        'misc sponsors';
      justify-content: center;
    }
  }

  .section-services {
    grid-area: services;
  }

  .section-client {
    grid-area: client;
  }

  .section-misc {
    grid-area: misc;
  }

  .section-sponsors {
    grid-area: sponsors;
  }

  & .vn-checkbox,
  & .vn-switch {
    grid-template-columns: min-content;
  }
}
</style>
