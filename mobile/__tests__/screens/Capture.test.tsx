/**
 * Capture screen tests
 *
 * Validates:
 * - Z8-AC04: Photo capture → pipeline flow
 * - UI: camera area, page thumbnails, gallery button, submit
 */
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react-native';

jest.mock('expo-router', () => ({
  router: { push: jest.fn(), replace: jest.fn() },
}));

jest.mock('react-native-safe-area-context', () => ({
  useSafeAreaInsets: () => ({ top: 44, bottom: 34, left: 0, right: 0 }),
}));

import CaptureScreen from '@/app/(tabs)/capture';

describe('Capture Screen', () => {
  it('shows capture header with page counter', () => {
    render(<CaptureScreen />);
    expect(screen.getByText('Capturer un cours')).toBeTruthy();
    expect(screen.getByText('0/30 pages')).toBeTruthy();
  });

  it('shows camera area with instruction', () => {
    render(<CaptureScreen />);
    expect(screen.getByText(/Cadre ton cahier/)).toBeTruthy();
  });

  it('shows photo capture button', () => {
    render(<CaptureScreen />);
    expect(screen.getByText(/Prendre une photo/)).toBeTruthy();
  });

  it('shows gallery button', () => {
    render(<CaptureScreen />);
    expect(screen.getByText(/Galerie/)).toBeTruthy();
  });

  it('shows finish button', () => {
    render(<CaptureScreen />);
    expect(screen.getByText(/Terminer/)).toBeTruthy();
  });

  describe('Page capture flow', () => {
    it('increments page count after capture', () => {
      render(<CaptureScreen />);
      fireEvent.press(screen.getByText(/Prendre une photo/));
      expect(screen.getByText('1/30 pages')).toBeTruthy();
    });

    it('shows thumbnail strip after first capture', () => {
      render(<CaptureScreen />);
      fireEvent.press(screen.getByText(/Prendre une photo/));
      expect(screen.getByText('Pages capturées')).toBeTruthy();
      // Thumbnail shows page number
      expect(screen.getByText('1')).toBeTruthy();
    });

    it('adds multiple page thumbnails', () => {
      render(<CaptureScreen />);
      fireEvent.press(screen.getByText(/Prendre une photo/));
      fireEvent.press(screen.getByText(/Prendre une photo/));
      fireEvent.press(screen.getByText(/Prendre une photo/));
      expect(screen.getByText('3/30 pages')).toBeTruthy();
    });

    it('submit navigates to processing screen', () => {
      const { router } = require('expo-router');
      render(<CaptureScreen />);
      // Capture at least one page first
      fireEvent.press(screen.getByText(/Prendre une photo/));
      fireEvent.press(screen.getByText(/Terminer/));
      expect(router.push).toHaveBeenCalledWith('/processing');
    });
  });
});
