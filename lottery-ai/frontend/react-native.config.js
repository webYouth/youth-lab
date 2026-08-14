module.exports = {
  dependencies: {
    expo: {
      platforms: {
        android: {
          packageImportPath: 'import expo.modules.ExpoModulesPackage;',
        },
      },
    },
    // 命中率图只用 SVG 渲染，跳过 Skia 原生链接（避免 peer/版本冲突）
    '@shopify/react-native-skia': {
      platforms: {
        android: null,
        ios: null,
      },
    },
  },
};
