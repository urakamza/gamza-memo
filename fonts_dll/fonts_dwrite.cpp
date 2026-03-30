#include <windows.h>
#include <dwrite.h>
#include <string>
#include <vector>

#pragma comment(lib, "dwrite.lib")

struct FontWeight {
    int weight;
    bool italic;
};

struct FontFamily {
    std::wstring family;
    std::vector<FontWeight> fonts;
};

// JSON 이스케이프
static std::wstring escapeJson(const std::wstring& s) {
    std::wstring out;
    for (wchar_t c : s) {
        if (c == L'"') out += L"\\\"";
        else if (c == L'\\') out += L"\\\\";
        else out += c;
    }
    return out;
}

extern "C" __declspec(dllexport)
int GetFontListJSON(wchar_t* buffer, int bufferSize) {
    IDWriteFactory* factory = nullptr;
    HRESULT hr = DWriteCreateFactory(
        DWRITE_FACTORY_TYPE_SHARED,
        __uuidof(IDWriteFactory),
        reinterpret_cast<IUnknown**>(&factory)
    );
    if (FAILED(hr) || !factory) return 0;

    IDWriteFontCollection* collection = nullptr;
    hr = factory->GetSystemFontCollection(&collection, FALSE);
    if (FAILED(hr) || !collection) {
        factory->Release();
        return 0;
    }

    std::wstring json = L"[";
    UINT32 familyCount = collection->GetFontFamilyCount();

    for (UINT32 i = 0; i < familyCount; i++) {
        IDWriteFontFamily* family = nullptr;
        if (FAILED(collection->GetFontFamily(i, &family))) continue;

        IDWriteLocalizedStrings* names = nullptr;
        if (FAILED(family->GetFamilyNames(&names))) {
            family->Release();
            continue;
        }

        // 한국어 이름 우선, 없으면 영문
        UINT32 index = 0;
        BOOL exists = FALSE;
        names->FindLocaleName(L"ko-kr", &index, &exists);
        if (!exists) {
            names->FindLocaleName(L"en-us", &index, &exists);
            if (!exists) index = 0;
        }

        UINT32 length = 0;
        names->GetStringLength(index, &length);
        std::vector<wchar_t> nameBuffer(length + 1);
        names->GetString(index, nameBuffer.data(), length + 1);
        std::wstring familyName(nameBuffer.data());

        names->Release();

        // weight 목록
        UINT32 fontCount = family->GetFontCount();
        std::wstring fontsJson = L"[";

        for (UINT32 j = 0; j < fontCount; j++) {
            IDWriteFont* font = nullptr;
            if (FAILED(family->GetFont(j, &font))) continue;

            int weight = (int)font->GetWeight();
            int stretch = (int)font->GetStretch();
            bool italic = (font->GetStyle() != DWRITE_FONT_STYLE_NORMAL);

            // face 이름 가져오기
            IDWriteLocalizedStrings* faceNames = nullptr;
            std::wstring faceName = L"";
            if (SUCCEEDED(font->GetFaceNames(&faceNames))) {
                UINT32 index = 0;
                BOOL exists = FALSE;
                faceNames->FindLocaleName(L"ko-kr", &index, &exists);
                if (!exists) {
                    faceNames->FindLocaleName(L"en-us", &index, &exists);
                    if (!exists) index = 0;
                }
                UINT32 length = 0;
                faceNames->GetStringLength(index, &length);
                std::vector<wchar_t> nameBuffer(length + 1);
                faceNames->GetString(index, nameBuffer.data(), length + 1);
                faceName = nameBuffer.data();
                faceNames->Release();
            }

            font->Release();

            if (j > 0) fontsJson += L",";
            fontsJson += L"{\"weight\":" + std::to_wstring(weight) +
                        L",\"italic\":" + (italic ? L"true" : L"false") +
                        L",\"stretch\":" + std::to_wstring(stretch) +
                        L",\"name\":\"" + escapeJson(faceName) + L"\"}";
        }
        fontsJson += L"]";

        if (i > 0) json += L",";
        json += L"{\"family\":\"" + escapeJson(familyName) + L"\",\"fonts\":" + fontsJson + L"}";

        family->Release();
    }

    json += L"]";

    collection->Release();
    factory->Release();

    if ((int)json.size() >= bufferSize) return 0;
    wmemcpy(buffer, json.c_str(), json.size() + 1);
    return 1;
}

BOOL APIENTRY DllMain(HMODULE, DWORD, LPVOID) {
    return TRUE;
}