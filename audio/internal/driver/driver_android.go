// Copyright 2016 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package driver

/*

#cgo LDFLAGS: -llog

#include <android/log.h>
#include <jni.h>
#include <stdlib.h>

// __android_log_print(ANDROID_LOG_ERROR, "NativeCode", "foo", "bar");

static char* initAudioTrack(uintptr_t java_vm, uintptr_t jni_env, jobject context,
    int sampleRate, int channelNum, int bytesPerSample, jobject* audioTrack) {
  JavaVM* vm = (JavaVM*)java_vm;
  JNIEnv* env = (JNIEnv*)jni_env;

  const jclass android_media_AudioFormat =
      (*env)->FindClass(env, "android/media/AudioFormat");
  const jclass android_media_AudioManager =
      (*env)->FindClass(env, "android/media/AudioManager");
  const jclass android_media_AudioTrack =
      (*env)->FindClass(env, "android/media/AudioTrack");

  const jint android_media_AudioManager_STREAM_MUSIC =
      (*env)->GetStaticIntField(
          env, android_media_AudioManager,
          (*env)->GetStaticFieldID(env, android_media_AudioManager, "STREAM_MUSIC", "I"));
  const jint android_media_AudioTrack_MODE_STREAM =
      (*env)->GetStaticIntField(
          env, android_media_AudioTrack,
          (*env)->GetStaticFieldID(env, android_media_AudioTrack, "MODE_STREAM", "I"));
  const jint android_media_AudioFormat_CHANNEL_OUT_MONO =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "CHANNEL_OUT_MONO", "I"));
  const jint android_media_AudioFormat_CHANNEL_OUT_STEREO =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "CHANNEL_OUT_STEREO", "I"));
  const jint android_media_AudioFormat_ENCODING_PCM_8BIT =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "ENCODING_PCM_8BIT", "I"));
  const jint android_media_AudioFormat_ENCODING_PCM_16BIT =
      (*env)->GetStaticIntField(
          env, android_media_AudioFormat,
          (*env)->GetStaticFieldID(env, android_media_AudioFormat, "ENCODING_PCM_16BIT", "I"));

  jint channel = android_media_AudioFormat_CHANNEL_OUT_MONO;
  switch (channelNum) {
  case 1:
    channel = android_media_AudioFormat_CHANNEL_OUT_MONO;
    break;
  case 2:
    channel = android_media_AudioFormat_CHANNEL_OUT_STEREO;
    break;
  default:
    return "invalid channel";
  }

  jint encoding = android_media_AudioFormat_ENCODING_PCM_8BIT;
  switch (bytesPerSample) {
  case 1:
    channel = android_media_AudioFormat_ENCODING_PCM_8BIT;
    break;
  case 2:
    channel = android_media_AudioFormat_ENCODING_PCM_16BIT;
    break;
  default:
    return "invalid bytesPerSample";
  }

  const jint bufferSize =
      (*env)->CallStaticIntMethod(
          env, android_media_AudioTrack,
          (*env)->GetStaticMethodID(env, android_media_AudioTrack, "getMinBufferSize", "(III)I"),
          sampleRate, channel, encoding);

  const jobject tmpAudioTrack =
      (*env)->NewObject(
          env, android_media_AudioTrack,
          (*env)->GetMethodID(env, android_media_AudioTrack, "<init>", "(IIIIII)V"),
          android_media_AudioManager_STREAM_MUSIC,
          sampleRate, channel, encoding, bufferSize,
          android_media_AudioTrack_MODE_STREAM);
  // Note that *audioTrack will never be released.
  *audioTrack = (*env)->NewGlobalRef(env, tmpAudioTrack);

  (*env)->CallVoidMethod(
      env, *audioTrack,
      (*env)->GetMethodID(env, android_media_AudioTrack, "play", "()V"));

  return NULL;
}

static char* writeToAudioTrack(uintptr_t java_vm, uintptr_t jni_env, jobject context,
    jobject audioTrack, void* data, int length) {
  // TODO: audioTrack object is not valid here?
  JavaVM* vm = (JavaVM*)java_vm;
  JNIEnv* env = (JNIEnv*)jni_env;

  const jclass android_media_AudioTrack =
      (*env)->FindClass(env, "android/media/AudioTrack");
  const jint android_media_AudioTrack_WRITE_NON_BLOCKING =
      (*env)->GetStaticIntField(
          env, android_media_AudioTrack,
          (*env)->GetStaticFieldID(env, android_media_AudioTrack, "WRITE_NON_BLOCKING", "I"));

  jbyteArray arr = (*env)->NewByteArray(env, length);
  (*env)->SetByteArrayRegion(env, arr, 0, length, data);

  jint result =
      (*env)->CallIntMethod(
          env, audioTrack,
          (*env)->GetMethodID(env, android_media_AudioTrack, "write", "([BIII)I"),
          arr, 0, length, android_media_AudioTrack_WRITE_NON_BLOCKING);
  // TODO: Check the result.
  return NULL;
}

*/
import "C"

import (
	"errors"
	"unsafe"
)

type Player struct {
	sampleRate     int
	channelNum     int
	bytesPerSample int
	audioTrack     C.jobject
}

func NewPlayer(sampleRate, channelNum, bytesPerSample int) (*Player, error) {
	p := &Player{
		sampleRate:     sampleRate,
		channelNum:     channelNum,
		bytesPerSample: bytesPerSample,
	}
	if err := runOnJVM(func(vm, env, ctx uintptr) error {
		audioTrack := C.jobject(nil)
		if msg := C.initAudioTrack(C.uintptr_t(vm), C.uintptr_t(env), C.jobject(ctx),
			C.int(sampleRate), C.int(channelNum), C.int(bytesPerSample),
			&audioTrack); msg != nil {
			return errors.New(C.GoString(msg))
		}
		p.audioTrack = audioTrack
		return nil
	}); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Player) Proceed(data []byte) error {
	if err := runOnJVM(func(vm, env, ctx uintptr) error {
		if msg := C.writeToAudioTrack(C.uintptr_t(vm), C.uintptr_t(env), C.jobject(ctx),
			p.audioTrack, unsafe.Pointer(&data[0]), C.int(len(data))); msg != nil {
			return errors.New(C.GoString(msg))
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (p *Player) Close() error {
	return nil
}
