import { useState } from 'react';
import {
  StyleSheet,
  ScrollView,
  Pressable,
  Image,
} from 'react-native';
import { View as RNView, Text as RNText } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { router } from 'expo-router';

import { useColors, Card, Button } from '@/components/Themed';
import { typography, spacing, radius } from '@/constants/Typography';

type CapturedPage = {
  id: string;
  uri: string;
};

export default function CaptureScreen() {
  const colors = useColors();
  const insets = useSafeAreaInsets();
  const [pages, setPages] = useState<CapturedPage[]>([]);
  const [subject, setSubject] = useState('');

  const handleCapture = () => {
    // TODO: integrate expo-camera
    const fakeId = String(pages.length + 1);
    setPages((prev) => [...prev, { id: fakeId, uri: '' }]);
  };

  const handleSubmit = () => {
    if (pages.length === 0) return;
    // TODO: call uploadPages API
    router.push('/processing');
  };

  return (
    <RNView style={[styles.container, { backgroundColor: colors.background, paddingTop: insets.top }]}>
      {/* Header */}
      <RNView style={styles.header}>
        <RNText style={[typography.h2, { color: colors.text }]}>
          Capturer un cours
        </RNText>
        <RNText style={[typography.caption, { color: colors.textSecondary, marginTop: spacing.xs }]}>
          {pages.length}/30 pages
        </RNText>
      </RNView>

      {/* Camera area */}
      <RNView style={[styles.cameraArea, { backgroundColor: colors.backgroundSecondary, borderColor: colors.border }]}>
        <RNText style={{ fontSize: 48 }}>📷</RNText>
        <RNText style={[typography.body, { color: colors.textSecondary, marginTop: spacing.md, textAlign: 'center' }]}>
          Cadre ton cahier{'\n'}dans le rectangle
        </RNText>
        <Button
          title="📸  Prendre une photo"
          variant="primary"
          onPress={handleCapture}
          style={{ marginTop: spacing.lg }}
        />
      </RNView>

      {/* Thumbnails */}
      {pages.length > 0 && (
        <RNView style={styles.thumbnailSection}>
          <RNText style={[typography.captionBold, { color: colors.textSecondary, marginBottom: spacing.sm }]}>
            Pages capturées
          </RNText>
          <ScrollView horizontal showsHorizontalScrollIndicator={false} style={{ gap: spacing.sm }}>
            {pages.map((page, idx) => (
              <RNView
                key={page.id}
                style={[styles.thumbnail, { backgroundColor: colors.tintLight, borderColor: colors.border }]}
              >
                <RNText style={[typography.bodyBold, { color: colors.tint }]}>{idx + 1}</RNText>
              </RNView>
            ))}
            <Pressable
              onPress={handleCapture}
              style={[styles.thumbnail, { backgroundColor: colors.backgroundSecondary, borderColor: colors.border, borderStyle: 'dashed' }]}
            >
              <RNText style={[typography.h3, { color: colors.textSecondary }]}>+</RNText>
            </Pressable>
          </ScrollView>
        </RNView>
      )}

      {/* Actions */}
      <RNView style={[styles.actions, { paddingBottom: insets.bottom + spacing.md }]}>
        <Button
          title="🖼  Galerie"
          variant="outline"
          onPress={() => {/* TODO: image picker */}}
          style={{ flex: 1 }}
        />
        <Button
          title="Terminer →"
          variant="primary"
          onPress={handleSubmit}
          style={{ flex: 1, opacity: pages.length === 0 ? 0.5 : 1 }}
        />
      </RNView>
    </RNView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1 },
  header: { paddingHorizontal: spacing.md, paddingVertical: spacing.md },
  cameraArea: {
    flex: 1,
    marginHorizontal: spacing.md,
    borderRadius: radius.lg,
    borderWidth: 2,
    borderStyle: 'dashed',
    alignItems: 'center',
    justifyContent: 'center',
  },
  thumbnailSection: { paddingHorizontal: spacing.md, paddingVertical: spacing.md },
  thumbnail: {
    width: 60,
    height: 80,
    borderRadius: radius.sm,
    borderWidth: 1,
    alignItems: 'center',
    justifyContent: 'center',
    marginRight: spacing.sm,
  },
  actions: {
    flexDirection: 'row',
    gap: spacing.sm,
    paddingHorizontal: spacing.md,
    paddingTop: spacing.md,
  },
});
