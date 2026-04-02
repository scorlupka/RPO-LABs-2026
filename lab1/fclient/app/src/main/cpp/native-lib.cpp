#include <jni.h>
#include <string>
#include <android/log.h>
#include <C:/Users/User/AndroidStudioProjects/libs/mbedtls/mbedtls/include/mbedtls/des.h>
#include <C:/Users/User/AndroidStudioProjects/libs/spdlog/spdlog/include/spdlog/spdlog.h>
#include <C:/Users/User/AndroidStudioProjects/libs/spdlog/spdlog/include/spdlog/sinks/android_sink.h>
#include <C:/Users/User/AndroidStudioProjects/libs/mbedtls/mbedtls/include/mbedtls/entropy.h>
#include <C:/Users/User/AndroidStudioProjects/libs/mbedtls/mbedtls/include/mbedtls/ctr_drbg.h>
#include <C:/Users/User/AndroidStudioProjects/libs/mbedtls/mbedtls/include/mbedtls/des.h>


mbedtls_entropy_context entropy;
mbedtls_ctr_drbg_context ctr_drbg;
const char *personalization = "fclient-sample-app";

#define LOG_INFO(...) __android_log_print(ANDROID_LOG_INFO, "fclient_ndk", __VA_ARGS__)

#define SLOG_INFO(...) android_logger->info( __VA_ARGS__ )
auto android_logger = spdlog::android_logger_mt("android", "fclient_ndk");

extern "C" JNIEXPORT jstring JNICALL
Java_com_example_fclient_MainActivity_stringFromJNI(JNIEnv* env, jobject /* this */) {
    std::string hello = "Hello Pasha from C++";
    LOG_INFO("Hello from c++ %d", 2023);
    SLOG_INFO("Hello from spdlog {0}", 2023);
    return env->NewStringUTF(hello.c_str());
}

extern "C" JNIEXPORT void JNICALL
Java_com_example_fclient_MainActivity_logForLibraries(JNIEnv* env, jclass /* clazz */) {
    // Для статического метода второй параметр - это jclass (класс), а не jobject (экземпляр)
    LOG_INFO("All libraries are good! %d", 2023);
}

extern "C" JNIEXPORT void JNICALL
Java_com_example_fclient_MainActivity_logForInt(JNIEnv* env, jclass /* clazz */, jint num) {
    LOG_INFO("Number is %d", num);
}

extern "C" JNIEXPORT jint JNICALL
Java_com_example_fclient_MainActivity_initRng(JNIEnv *env, jclass clazz) {
    mbedtls_entropy_init( &entropy );
    mbedtls_ctr_drbg_init( &ctr_drbg );

    return mbedtls_ctr_drbg_seed( &ctr_drbg , mbedtls_entropy_func, &entropy,
                                  (const unsigned char *) personalization,
                                  strlen( personalization ) );
}

extern "C" JNIEXPORT jbyteArray JNICALL
Java_com_example_fclient_MainActivity_randomBytes(JNIEnv *env, jclass, jint no) {
    uint8_t * buf = new uint8_t [no];
    mbedtls_ctr_drbg_random(&ctr_drbg, buf, no);
    jbyteArray rnd = env->NewByteArray(no);
    env->SetByteArrayRegion(rnd, 0, no, (jbyte *)buf);
    delete[] buf;
    return rnd;
}

// https://tls.mbed.org/api/des_8h.html
extern "C" JNIEXPORT jbyteArray JNICALL
Java_com_example_fclient_MainActivity_encrypt(JNIEnv *env, jclass, jbyteArray key, jbyteArray data)
{
    jsize ksz = env->GetArrayLength(key);
    jsize dsz = env->GetArrayLength(data);
    if ((ksz != 16) || (dsz <= 0)) {
        return env->NewByteArray(0);
    }
    mbedtls_des3_context ctx;
    mbedtls_des3_init(&ctx);

    jbyte * pkey = env->GetByteArrayElements(key, 0);

    // Паддинг PKCS#5
    int rst = dsz % 8;
    int sz = dsz + 8 - rst;
    uint8_t * buf = new uint8_t[sz];
    for (int i = 7; i > rst; i--)
        buf[dsz + i] = rst;
    jbyte * pdata = env->GetByteArrayElements(data, 0);
    std::copy(pdata, pdata + dsz, buf);
    mbedtls_des3_set2key_enc(&ctx, (uint8_t *)pkey);
    int cn = sz / 8;
    for (int i = 0; i < cn; i++)
        mbedtls_des3_crypt_ecb(&ctx, buf + i*8, buf + i*8);
    jbyteArray dout = env->NewByteArray(sz);
    env->SetByteArrayRegion(dout, 0, sz, (jbyte *)buf);
    delete[] buf;
    env->ReleaseByteArrayElements(key, pkey, 0);
    env->ReleaseByteArrayElements(data, pdata, 0);
    return dout;
}

extern "C" JNIEXPORT jbyteArray JNICALL
Java_com_example_fclient_MainActivity_decrypt(JNIEnv *env, jclass, jbyteArray key, jbyteArray data)
{
    jsize ksz = env->GetArrayLength(key);
    jsize dsz = env->GetArrayLength(data);
    if ((ksz != 16) || (dsz <= 0) || ((dsz % 8) != 0)) {
        return env->NewByteArray(0);
    }
    mbedtls_des3_context ctx;
    mbedtls_des3_init(&ctx);

    jbyte * pkey = env->GetByteArrayElements(key, 0);

    uint8_t * buf = new uint8_t[dsz];

    jbyte * pdata = env->GetByteArrayElements(data, 0);
    std::copy(pdata, pdata + dsz, buf);
    mbedtls_des3_set2key_dec(&ctx, (uint8_t *)pkey);
    int cn = dsz / 8;
    for (int i = 0; i < cn; i++)
        mbedtls_des3_crypt_ecb(&ctx, buf + i*8, buf +i*8);

    //PKCS#5. упрощено. по соображениям безопасности надо проверить каждый байт паддинга
    int sz = dsz - 8 + buf[dsz-1];


    jbyteArray dout = env->NewByteArray(sz);
    env->SetByteArrayRegion(dout, 0, sz, (jbyte *)buf);
    delete[] buf;
    env->ReleaseByteArrayElements(key, pkey, 0);
    env->ReleaseByteArrayElements(data, pdata, 0);
    return dout;
}

extern "C" JNIEXPORT void JNICALL
Java_com_example_fclient_MainActivity_logByteArray( JNIEnv *env, jclass clazz, jbyteArray array, jstring label)
{
    if (array == NULL) {
        LOG_INFO("logByteArray: array is NULL");
        return;
    }

    const char* label_str = (label != NULL) ? env->GetStringUTFChars(label, NULL) : "byte[]";

    jsize len = env->GetArrayLength(array);
    LOG_INFO("%s → length = %d", label_str, (int)len);

    if (len == 0) {
        if (label != NULL) env->ReleaseStringUTFChars(label, label_str);
        return;
    }

    jbyte* bytes = env->GetByteArrayElements(array, NULL);
    if (bytes == NULL) {
        LOG_INFO("%s → GetByteArrayElements failed", label_str);
        if (label != NULL) env->ReleaseStringUTFChars(label, label_str);
        return;
    }

    // Выводим по 16 байт в строке в hex-формате
    const int BYTES_PER_LINE = 16;
    char hex_line[BYTES_PER_LINE * 3 + 1];  // 2 символа + пробел + \0

    for (int i = 0; i < len; i += BYTES_PER_LINE) {
        int line_len = (i + BYTES_PER_LINE < len) ? BYTES_PER_LINE : (len - i);

        int pos = 0;
        for (int j = 0; j < line_len; j++) {
            unsigned char b = (unsigned char)bytes[i + j];
            sprintf(hex_line + pos, "%02X ", b);
            pos += 3;
        }
        hex_line[pos] = '\0';

        LOG_INFO("%s  %04d | %s", label_str, i, hex_line);
    }

    env->ReleaseByteArrayElements(array, bytes, JNI_ABORT);

    if (label != NULL) env->ReleaseStringUTFChars(label, label_str);
}


JavaVM* gJvm = nullptr;

JNIEXPORT jint JNICALL JNI_OnLoad (JavaVM* pjvm, void* reserved)
{
    gJvm = pjvm;
    return JNI_VERSION_1_6;
}

JNIEnv* getEnv (bool& detach)
{
    JNIEnv* env = nullptr;
    int status = gJvm->GetEnv ((void**)&env, JNI_VERSION_1_6);
    detach = false;
    if (status == JNI_EDETACHED)
    {
        status = gJvm->AttachCurrentThread (&env, NULL);
        if (status < 0)
        {
            return nullptr;
        }
        detach = true;
    }
    return env;
}

void releaseEnv (bool detach, JNIEnv* env)
{
    if (detach && (gJvm != nullptr))
    {
        gJvm->DetachCurrentThread ();
    }
}



extern "C"
JNIEXPORT jboolean JNICALL
Java_com_example_fclient_MainActivity_transaction(JNIEnv *xenv, jobject xthiz, jbyteArray xtrd) {
    // Создаём глобальные ссылки, чтобы объекты жили после выхода из функции
    jobject thiz = xenv->NewGlobalRef(xthiz);
    jbyteArray trd = (jbyteArray) xenv->NewGlobalRef(xtrd);

    // Запускаем фоновый поток
    std::thread t([thiz, trd]() {
        bool detach = false;
        JNIEnv *env = getEnv(detach);

        if (env == nullptr) {
            LOG_INFO("transaction thread: failed to get env");
            releaseEnv(detach, env);
            return;
        }

        // Получаем класс и методы
        jclass cls = env->GetObjectClass(thiz);
        if (cls == nullptr) {
            LOG_INFO("GetObjectClass failed");
            releaseEnv(detach, env);
            return;
        }

        jmethodID enterPinId = env->GetMethodID(cls, "enterPin", "(ILjava/lang/String;)Ljava/lang/String;");
        jmethodID resultId   = env->GetMethodID(cls, "transactionResult", "(Z)V");

        if (enterPinId == nullptr || resultId == nullptr) {
            LOG_INFO("GetMethodID failed");
            releaseEnv(detach, env);
            return;
        }

        // Парсим TRD → сумма
        jbyte *p = env->GetByteArrayElements(trd, nullptr);
        jsize sz = env->GetArrayLength(trd);

        if (sz != 9 || p[0] != (jbyte)0x9F || p[1] != (jbyte)0x02 || p[2] != (jbyte)0x06) {
            LOG_INFO("Invalid TRD format: len=%d, %02X %02X %02X", sz, (uint8_t)p[0], (uint8_t)p[1], (uint8_t)p[2]);
            env->ReleaseByteArrayElements(trd, p, JNI_ABORT);
            env->CallVoidMethod(thiz, resultId, JNI_FALSE);
            releaseEnv(detach, env);
            return;
        }

        char buf[13];
        for (int i = 0; i < 6; i++) {
            uint8_t n = (uint8_t)p[3 + i];
            buf[i*2]     = ((n >> 4) & 0x0F) + '0';
            buf[i*2 + 1] = (n & 0x0F) + '0';
        }
        buf[12] = '\0';

        jstring jamount = env->NewStringUTF(buf);
        env->ReleaseByteArrayElements(trd, p, JNI_ABORT);

        int ptc = 3;
        jboolean success = JNI_FALSE;

        while (ptc > 0) {
            jstring jpin = (jstring) env->CallObjectMethod(thiz, enterPinId, ptc, jamount);

            if (jpin == nullptr) {
                // Пользователь отменил
                break;
            }

            const char *utf = env->GetStringUTFChars(jpin, nullptr);
            if (utf != nullptr && strcmp(utf, "1234") == 0) {
                success = JNI_TRUE;
            }

            if (utf) env->ReleaseStringUTFChars(jpin, utf);
            env->DeleteLocalRef(jpin);

            if (success) break;
            ptc--;
        }

        env->DeleteLocalRef(jamount);

        // Сообщаем результат в Java
        env->CallVoidMethod(thiz, resultId, success);

        // Освобождаем глобальные ссылки
        env->DeleteGlobalRef(thiz);
        env->DeleteGlobalRef(trd);

        releaseEnv(detach, env);
    });

    t.detach();

    return JNI_TRUE;
}



